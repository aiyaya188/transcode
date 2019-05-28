package controlers

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	//"strings"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"../cache"
	"../common"
	"../definition"
	"../models"
)

/*
type Transition struct {
	IntFile     *os.File  //写入文件
	OutFile     *os.File  //输出文件
	Duration    int       //视频总时长
	CurrentTime int       //当前转码时长
	Status      chan bool //转码状态
	DeBug       bool      //是否开启debug
}
*/
/*
func NewFunc(intFile *os.File) *Transition {
	t := new(Transition)
	t.IntFile = intFile
	t.Status = make(chan bool)
	t.DeBug = true
	return t
}
*/
type WorkingCount struct {
	count int
	lock  sync.RWMutex
}

var Workings WorkingCount

func init() {
	Workings.count = 0
}

func (w *WorkingCount) addWorkingCount() {
	w.lock.Lock()
	w.count = w.count + 1
	w.lock.Unlock()
}

func (w *WorkingCount) delWorkingCount() {
	w.lock.Lock()
	w.count = w.count - 1
	w.lock.Unlock()
}

func (w *WorkingCount) GetValue() int {
	w.lock.Lock()
	res := w.count
	w.lock.Unlock()
	return res
}

//死循环调用GetJob
func TranCodeRun() {
	for {
		//common.Log("开始调度转码任务-----------------------------")
		go GetJob()
		time.Sleep(2 * time.Second)
	}
}

//GetJob 从缓存列表获取等待转码的文件路径
func GetJob() {
	currentJobs := GetCurrentJobs()
	common.Log2("当前任务数量：", currentJobs)
	maxJobs := GetMaxJobs()
	common.Log2("最大数量：", maxJobs)
	if currentJobs >= maxJobs {
		common.Log("jobs is full")
		return
	}
	listKey := definition.KeyTranJob
	var jobValue string
	jobValue = cache.PopList(listKey)
	if jobValue == "" {
		jobValue = FullJobList()
	}
	if jobValue != "" {
		//common.Log("jobvalue:", jobValue)
		startTranCode(jobValue)
		//视频状态更新为成功
	} else {
		//common.Log("当前没有任务")
		return
	}
}

//返回m3u地址
func startTranCode(job string) {
	Workings.addWorkingCount()
	currentJobs := GetCurrentJobs()
	common.Log2("startTranCode当前任务数量：", currentJobs)
	var video definition.Videos
	var bitrate int
	//var videoDuration int
	json.Unmarshal([]byte(job), &video)
	//var m3u8Path string
	path := common.GetConfig("system", "storageSource").String() + video.Path
	path1 := common.GetConfig("system", "storageDest").String() + video.DestPath
	common.MkdirForSource(path1)
	models.UpdateStatus(video.TaskId, definition.VideTranCoding)
	common.Log2("path2:", path1)
	common.Log2("taskid:", video.TaskId)
	//bitrate, duration, res := getVideoInfo(path, video.TaskId)
	bitrate, videoLen, streams, res := getVideoInfo(path, video.TaskId)
	if res == false {
		return
	}
	waterRes := WaterVideo(path, path1, video.TaskId, bitrate, videoLen, streams)
	defer Workings.delWorkingCount()
	if waterRes == "ok" {
		putS3(video.DestPath, video.TaskId)

	}
	return
}

/**
开发调试信息
*/
/*
func (t *Transition) SetDeBug(b bool) *Transition {
	t.DeBug = b
	return t
}
*/
/**
时间解析
*/
func timeEncode(t string) int {
	time := strings.Trim(t, " ")
	common.Log2("time:", time)
	if time == `N/A` {
		return 0
	}
	hour, _ := strconv.Atoi(time[:2])
	minute, _ := strconv.Atoi(time[3:5])
	second, _ := strconv.Atoi(time[6:8])
	return second + minute*60 + hour*3600
}

func AddPrenameToPath(path string, prename string) string {
	sourcePath := path
	place := strings.LastIndex(sourcePath, "/")
	prePath := sourcePath[:place+1]
	fixPath := sourcePath[place+1:]
	destPath := prePath + prename + fixPath
	return destPath
}

func cutFixname(path string) string {
	sourcePath := path
	place := strings.LastIndex(sourcePath, ".")
	newPath := sourcePath[:place]
	return newPath
}

func getVideoInfo(s string, taskID string) (int, int, int, bool) {
	var videoDuration int
	var bitrate int
	var streams int
	videoDuration = 0
	bitrate = 0
	streams = 2
	param := "ffmpeg -i " + s
	cmd := exec.Command("/bin/bash", "-c", param)
	stderrPipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		common.Log2("geti video info param :", param)
		common.Log2("Execute failed when Start:" + err.Error())
		return 0, 0, 0, false
	}
	common.Log2("test1")
	err_bytes, _ := ioutil.ReadAll(stderrPipe)
	common.Log2("test2")
	//reader := bufio.NewReader(stderrPipe)
	defer stderrPipe.Close()
	if err := cmd.Wait(); err != nil {
		//common.Log2("Execute failed when Wait:" + err.Error())
		//return 0, 0, false
	}
	//获取时长
	common.Log2("test3")
	reg1 := regexp.MustCompile(`Duration:(.*?),`)
	common.Log2("test4")
	snatch1 := reg1.FindStringSubmatch(string(err_bytes))
	common.Log2("test5")
	common.Log2("snatch1:", snatch1)
	//common.Log2("video len:",snatch1[1])
	if len(snatch1) > 1 {
		common.Log2("test6")
		videoDuration = timeEncode(snatch1[1])
		common.Log2("test7")
		models.UpdatePlayTime(taskID, snatch1[1])
		common.Log2("video len:", strings.TrimSpace(snatch1[1]))
		common.Log2("视频Duration: %d 秒\n", videoDuration)
	} else {
		common.Log2("没有时长信息\n")
	}
	//获取比特率
	reg8 := regexp.MustCompile(`Stream #0:(.*?)kb/s`)
	snatch8 := reg8.FindStringSubmatch(string(err_bytes))
	if len(snatch8) > 1 {
		place := strings.LastIndex(snatch8[1], ",")
		temp := snatch8[1]
		common.Log2("bitrate:", strings.TrimSpace(temp[place+1:]))
		var err error
		bitrate, err = strconv.Atoi(strings.TrimSpace(temp[place+1:]))
		if err == nil {
			common.Log2("bitrate:", bitrate)
		}
	} else {
		common.Log2("没有比特率信息\n")
	}
	common.Log2("test4")
	//获取音频流
	patten := `Stream #0:1`
	res1, _ := regexp.MatchString(patten, string(err_bytes))
	if res1 == false {
		streams = 1
		common.Log2("没有音频信息\n")
	}
	return bitrate, videoDuration, streams, true
}

func WaterVideo(path string, finalPath string, taskID string, bit int, videoLen int, streams int) string {
	sourcePath := path
	temp := strings.LastIndex(sourcePath, ".")
	temp1 := sourcePath[:temp] + ".mp4"
	common.Log2("temp1:", temp1)
	destPath := AddPrenameToPath(temp1, "water_")
	common.Log2("destPath:", destPath)

	waterPath := common.GetConfig("system", "waterPicture").String()
	//waterPath := "water.jpg"
	var waterPlace int
	waterPlace = models.GetWatermarkPlace()
	common.Log2("water place:", waterPlace)
	var overlay string

	overlay = `overlay=0:0`
	switch waterPlace {
	case 0: //左上
		overlay = `overlay=0:0`
		break
	case 2: //左下
		overlay = `overlay=0:main_h-overlay_h`
		break
	case 1: //右上
		overlay = `overlay=main_w-overlay_w:0`
		break
	case 3: //右下
		overlay = `overlay=main_w-overlay_w:main_h-overlay_h`
	}
	common.Log2("水印开始\n")
	param := `/opt/local/ffmpeg/bin/ffmpeg -i ` + sourcePath + ` -i ` + waterPath + " -filter_complex" + " " + overlay + " " + destPath
	common.Log2("parm : ", param)
	start := time.Now()
	_, err := exec.Command("bash", "-c", param).CombinedOutput()
	if err != nil {
		common.Log2("error: " + err.Error())
		common.Log2("水印失败\n")
		//taskID, _ := models.GetTaskID(videoSample)
		models.UpdateStatus(taskID, definition.VideoTranErr)
		return ""
	} else {
		common.Log2(" exec time  ", time.Now().Sub(start).Seconds())
	}
	common.Log2("水印完成\n")
	var zimuPath string
	var res bool
	res, zimuPath = ZimuHiProcess(destPath)
	if res == false {
		common.Log2("字幕失败\n")
		models.UpdateStatus(taskID, definition.VideoTranErr)
		return ""
	}
	var flag int
	if bit > 480 {
		flag = 1
		common.Log2("480双码率m3u8开始")
	} else {
		flag = 0
		common.Log2("250单码率m3u8开始")
	}

	res = M3u8Process(flag, zimuPath, finalPath, videoLen, streams, sourcePath)
	if res == false {
		common.Log2("切片失败\n")
		models.UpdateStatus(taskID, definition.VideoTranErr)
		return ""
	}
	common.Log2("m3u8切片结束")
	models.UpdateStatus(taskID, definition.VideoToCloud)
	return "ok"
}
func putS3(path string, md5 string) {
	//组合json数据
	place := strings.LastIndex(path, `/`)
	pathDir := path[:place]
	var cloudUpOrder definition.CloudUpOrder
	cloudUpOrder.Md5 = md5
	cloudUpOrder.Path = pathDir
	order, _ := json.Marshal(cloudUpOrder)
	cache.PushList(definition.CloudUpKey, string(order))
	return
}

//FullJobList : 从mysql获取等待转码的视频到队列,并且弹出第一个任务
func FullJobList() string {
	var videoSets []definition.Videos
	var videoCount int
	videoCount, videoSets = models.GetWaitingVideos()
	if videoCount == 0 {
		common.Log2("FullJobList:videocount =0")
		return ""
	} else {
		//videoJson,err:= json.Marshal(vide)
		for i := 0; i < videoCount; i++ {
			videoJson, err := json.Marshal(videoSets[i])
			if err != nil {
				common.Log2("video json fatal:" + err.Error())
				panic(err)
			}
			common.Log2("push video")
			cache.PushList(definition.KeyTranJob, string(videoJson))

		}
		return cache.PopList(definition.KeyTranJob)
	}
}

//GetMaxJobs : 从redis获取最大的工作进程数量，如果获取不到，就从mysql获取
func GetMaxJobs() int {
	/*
		temp := common.GetConfig("system", "maxJobs").String()
	*/
	temp := models.GetMaxJobs()
	maxJobs, _ := strconv.Atoi(temp)
	return maxJobs
}

//GetCurrentJobs
func GetCurrentJobs() int {
	res := Workings.GetValue()
	return res
}

func DelVideoProcess(w http.ResponseWriter, r *http.Request) {
	//RequestAli()
	if r.Method == "OPTIONS" {
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var delData definition.DelVideos
	common.Log3("delvideo:", b)
	err = json.Unmarshal(b, &delData)

	common.Log3("test11")
	if err != nil {
		common.Log3(":" + err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	common.Log3("del path:" + delData.Path)
	w.Header().Set("access-control-allow-origin", "*")  //允许访问所有域
	w.Header().Add("access-control-allow-headers", "*") //header的类型
	w.Header().Add("access-control-expose-headers", "*")
	w.Header().Set("content-type", "application/json")

	var responBlock definition.DelVideoStatus
	if delData.Path != "" {
		err2 := os.RemoveAll(delData.Path)
		common.Log3("delvideo:", b)
		if err2 == nil {
			responBlock.Status = "1"
		} else {
			responBlock.Status = "0"
		}
	} else {

		responBlock.Status = "0"
	}
	output, _ := json.Marshal(responBlock)
	w.Write(output)
	return
}

func CloudNotifyProcess(w http.ResponseWriter, r *http.Request) {
	//RequestAli()
	if r.Method == "OPTIONS" {
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var notify definition.CloudNotify
	common.Log3("cload notify:", b)
	err = json.Unmarshal(b, &notify)
	if err != nil {
		common.Log3("cloudnotify err:" + err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	if notify.Status == definition.SUCCESS {
		common.Log3(notify.Md5, "ok")
		models.UpdateStatus(notify.Md5, definition.VideSuccess)
	} else {
		common.Log3(notify.Md5, "fail")
		models.UpdateStatus(notify.Md5, definition.VideoS3Err)
	}
}

func ZimuHiProcess(path string) (bool, string) {
	sourcePath := path
	var posti string
	subtitle, color, size, startpost := models.GetSubtitle()
	if startpost == "1" {
		posti = ""
	}
	if startpost == "2" {
		posti = `:y=(h-th)/2`
	}
	if startpost == "3" {
		posti = `:y=h-th`
	}
	common.Log2("开始字幕")
	destPath := AddPrenameToPath(sourcePath, "zimu_")
	param := `/opt/local/ffmpeg/bin/ffmpeg -y -threads 2 -i ` + sourcePath + ` -filter:v` + ` drawtext="/usr/share/fonts/chinese/simsun.ttc: text=` + `'` + subtitle + `'` + ` :fontcolor=` + color + `:fontsize=` + size + posti + `:x=(tw-mod(8*n\,w+tw*2.8))"` + `  -codec:v libx264  -codec:a  copy -y` + " " + destPath + ""
	//param := `/opt/local/ffmpeg/bin/ffmpeg -y -threads 2 -i ` + sourcePath + ` -filter:v` + ` drawtext="/usr/share/fonts/chinese/simsun.ttc: text=` + subtitle + ` :fontcolor=` + color + `:fontsize=` + size + posti + `:x=(tw-mod(8*n\,w+tw*2.8))"` + `  -codec:v libx264  -codec:a  copy -y` + " " + destPath + ""
	common.Log2("parm : ", param)
	start := time.Now()
	_, err := exec.Command("bash", "-c", param).CombinedOutput()
	if err != nil {
		common.Log2("error: " + err.Error())
		return false, ""
	} else {
		common.Log2(" exec time  ", time.Now().Sub(start).Seconds())
	}
	common.Log2("字幕结束")
	return true, destPath
}

func M3u8Process(mode int, path string, finalPath string, videoLen int, streams int, source string) bool {
	sourcePath := path
	destPath := finalPath

	common.Log2("开始m3u8\n")
	var param string
	var newName string
	var newName1 string
	var newName2 string
	var mvParam string
	place := strings.LastIndex(destPath, "/")
	newName = destPath[place+1:]
	newName2 = destPath[:place]
	place1 := strings.LastIndex(newName, ".")
	newName1 = newName[:place1-1]
	var status bool
	status = true
	var ffcmd480Dual string
	var ffcmd480single string
	var ffcmd250Dual string
	// var ffcmd250single string

	var fragment int
	var fragmentStr string
	fragmentStr = models.GetFragmentDuartion()
	fragment, _ = strconv.Atoi(fragmentStr)
	if fragment > videoLen {
		fragmentStr = strconv.Itoa(videoLen)
	}
	if videoLen == 0 {
		fragmentStr = "2"
	}
	ffcmd480Dual = `ffmpeg -y -threads 2 -i ` + sourcePath + ` -b:v:0 480000  -bufsize 4M -b:v:1 250000 -maxrate 250000 -bufsize 4M  -map 0:v -map 0:a -map 0:v -map 0:a -f hls -hls_time ` + fragmentStr + ` -hls_flags split_by_time -hls_list_size 0  -var_stream_map "v:0,a:0 v:1,a:1" -master_pl_name ` + `'` + newName + `'` + ` ` + newName1 + `%v` + `.m3u8`
	ffcmd480single = `ffmpeg -y -threads 2 -i ` + sourcePath + ` -b:v:0 480000 -b:v:1 250000   -map 0:v  -map 0:v  -f hls -hls_time ` + fragmentStr + `  -hls_flags split_by_time -hls_list_size 0  -var_stream_map "v:0 v:1" -master_pl_name ` + `'` + newName + `'` + ` ` + newName1 + `%v` + `.m3u8`
	ffcmd250Dual = `ffmpeg -y -threads 2 -i ` + sourcePath + ` -b:v 250000 -f hls -hls_time ` + fragmentStr + ` -hls_flags split_by_time -hls_list_size 0  -hls_playlist_type vod ` + destPath
	if mode == 1 {
		//  newPath := cutFixname(destPath)
		if streams == 2 {
			param = ffcmd480Dual
		} else {
			param = ffcmd480single
		}
	} else {
		param = ffcmd250Dual
	}

	common.Log2("parm : ", param)
	//cmd := exec.Command("/opt/local/ffmpeg/bin/ffmpeg", "-y", "-threads","-2","-i", sourcePath, "-filter:v", drawtext, "-codec:v", "libx264", "-codec:a", "copy", "-y", destPath)
	start := time.Now()
	_, err := exec.Command("bash", "-c", param).CombinedOutput()
	if err != nil {
		common.Log2("error: " + err.Error())
		//common.Log2("try single")
		status = false
		/*
			if mode == 1 {
				param = ffcmd480single
				start1 := time.Now()
				_, err12 := exec.Command("bash", "-c", param).CombinedOutput()
				if err12 != nil {
					common.Log2("error: " + err12.Error())
				}
				status = 1
				common.Log2(" exec time  ", time.Now().Sub(start1).Seconds())
			}
		*/
	}
	common.Log2("status:", status)
	if status == true {

		mvParam = `mv ` + newName1 + `*` + ` ` + newName2
		common.Log2("mv param:", mvParam)
		_, err1 := exec.Command("bash", "-c", mvParam).CombinedOutput()
		if err1 != nil {
			common.Log2("error: " + err1.Error())
		}
		common.Log2(" exec time  ", time.Now().Sub(start).Seconds())
	}
	DeleteSource(path)
	//删除源文件
	/*
		placeRmName := strings.LastIndex(path, "_")
		rmName := path[placeRmName+1:]
		placeRmName1 := strings.LastIndex(path, "/")
		rmName1 := path[:placeRmName1]
		//zimuPath =   rmName1+`/`+`zimu_water_`+rmName
		waterPath := rmName1 + `/` + `water_` + rmName
		common.Log2("rm path:", path)
		os.Remove(path)
		common.Log2("rm path:", waterPath)
		os.Remove(waterPath)
		//  mp4Path := rmName1+`/`+rmName
		common.Log2("rm path:", source)
		os.Remove(source)
	*/
	return status
}
func DeleteSource(pathSource string) {
	place := strings.LastIndex(pathSource, `/`)
	path := pathSource[:place]
	// 创建文件夹
	err := os.RemoveAll(path)
	if err != nil {
		common.Log1("remove sorce err:", err)
	}
}
