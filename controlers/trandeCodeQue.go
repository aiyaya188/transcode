package controlers

/*
import (
	"sync"

	"../common"
	"../definition"
)

type WorkingCount struct {
	count int
	lock  sync.RWMutex
}

var workings WorkingCount
var queue common.ItemQueue

func initQueue() *common.ItemQueue {
	if queue.items == nil {
		queue = common.ItemQueue{}
		queue.New()
		return &queue
	}
	return &queue
}

func init() {
	workings.count = 0
	&queue = initQueue()
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

//死循环调用GetJob
func TranCodeRun() {
}

//GetJob 从缓存列表获取等待转码的文件路径
func GetJob() {
	currentJobs := GetCurrentJobs()
	maxJobs := GetMaxJobs()
	if currentJobs >= maxJobs {
		common.Log("jobs is full")
		return
	}
	listKey := definition.KeyTranJob
	var jobValue string
	jobValue = common.PopList(listKey)
	if jobValue == "" {
		jobValue = FullJobList()
	}
	if jobValue != "" {
		res := startTranCode(jobValue)
		putS3(res)
		//视频状态更新为成功
	} else {
		return
	}
}

//返回m3u地址
func startTranCode(path string) string {
	//当前任务+1
	//视频状态更新为正在转码
	//当前任务-1
	//视频状态更新为转码完成
	//计算时长

	return ""
}
func waterVideo(path string) {
	return
}
func zimuVideo(path string) {
	return
}
func tranMp4Video(path string) {
	return
}

func m3u8Video(path string) {
	return
}

func putS3(path string) {
	//视频状态更新为上传s3
}

//FullJobList : 从mysql获取等待转码的视频到队列,并且弹出第一个任务
func FullJobList() string {
	return ""
}

//GetMaxJobs : 从redis获取最大的工作进程数量，如果获取不到，就从mysql获取
func GetMaxJobs() int {
	return 0
}

//GetCurrentJobs
func GetCurrentJobs() int {
	return 0
}
*/
