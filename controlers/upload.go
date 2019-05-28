/**
* @file upload.go
* @brief: 上传控制器
* @author frankie@gmail.com
* @version v1.0
* @date 2018-10-22
 */

package controlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"../definition"
	"../models"

	//"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"../cache"
	"../common"
)

var blocksize int64 //Blocksize  分块长度

/* --------------------------------------------------------------------------*/
/**
* @brief: init 初始化上传功能需要的变量
*
* @returns:
 */
/* ----------------------------------------------------------------------------*/
func init() {
	temp := common.GetConfig("system", "blocksize").String()
	blocksize, _ = strconv.ParseInt(temp, 10, 64)
	common.Log1(fmt.Sprintf("blocksize is %n", blocksize))

	/*common.Log1("blocksize :"+temp)*/
}

//PathExists:  判断文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//mkdirForData: 根据日期创建目录

func mkdirForData(path1 string) {
	dateString := time.Now().Format("2006-01-02")
	path := path1 + dateString
	exist, err := PathExists(path)
	if err != nil {
		common.Log("get dir error")
		return
	}
	if exist {
		return
	} else {
		// 创建文件夹
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			common.Log("mkdir failed!")
		} else {
			common.Log(path + "mkdir success!")
		}
	}
}

/** writeChunk
* @brief writeChunk 写入文件块
*                   检查文件是否已完整，检查期望offset是否与写入一致
*                   写入文件，返回写入后的长度
*
* @param string
* @param int64
* @param
* @param bool
* @param error
*
* @returns  期望offset,是否完成，内部操作是否正确，错误码
 */
/* ----------------------------------------------------------------------------*/
func writeChunk(videoSample definition.Videos, src io.Reader) (int64, bool, bool, error) {
	res, videoPath := models.PreWrite(videoSample)
	//path := fmt.Sprintf(os.Getenv("USERVER_INI")+"/DATA/%s", videoPath)
	path := common.GetConfig("system", "storageSource").String() + videoPath
	//storagePath := common.GetConfig("system", "storageSource").String()
	common.Log("path:" + path)
	if res == false {
		common.Log("prwrite false")
		return 0, false, false, errors.New("prowrite fail")
	}
	//mkdirForData(storagePath)
	common.MkdirForSource(path)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		common.Log("open file fail")
		return 0, false, false, err
	}
	defer file.Close()

	common.Log("start copy")
	n, err := io.Copy(file, src)
	if err != nil {
		common.Log("copy data fail")
		return n, false, false, err
	}
	common.Log("copy data end")
	//更新期望偏移值到数据库,并返回期望偏移值
	videoSample.Offset = videoSample.Offset + n
	//taskID, _ := models.GetTaskID(videoSample)
	//common.Log("taskid:" + taskID)
	//models.UpdateOffset(taskID, videoSample.Offset)
	key := models.UpadateOffsetCache(videoSample, videoSample.Offset)
	if videoSample.FileSize <= videoSample.Offset {
		common.Log("upload complete")
		taskID, _ := models.GetTaskID(videoSample)
		models.UpdateStatus(taskID, definition.VideoWaitTran)
		//删除缓存
		cache.DelCache(key)
		return videoSample.Offset, true, true, err
	}
	return videoSample.Offset, false, true, err
}
func HandleUpload1(w http.ResponseWriter, r *http.Request) {
	//year, month, day := time.Date()
	//day := time.Day()
	common.Log(" handle uload")
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
	w.Header().Add("Access-Control-Allow-Headers", "tranType,offset,FileSize,fileMd5,FileType,FileName,playUrl,offset,blockSize,retcode,errmsg") //header的类型
	w.Header().Add("Access-Control-Expose-Headers", "tranType,offset,FileSize,fileMd5,FileType,FileName,playUrl,offset,blockSize,retcode,errmsg")
	if r.Method == "OPTIONS" {
		return
	}
	tranType := r.Header.Get("tranType")
	if tranType == "" {
		common.Log("trantype is nill")
		respond(w, 0, 0, definition.RetFailed, "", definition.Msg_TranType_Nill)
		return
	} else {
		common.Log1("tranType:", tranType)
		respond(w, 0, 0, definition.RetSucceed, "", definition.Msg_Ok)
		return
	}

}

/*HandleUpload --------------------------------------------------------------------------*/
/**
* @brief: HandleUpload 处理文件上传
*
* @param: http.ResponseWriter
* @param: http.Request
*
* @returns:
 */
/* ----------------------------------------------------------------------------*/
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	//year, month, day := time.Date()
	//day := time.Day()
	common.Log(" handle uload")
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
	//w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
	w.Header().Add("Access-Control-Allow-Headers", "upParam") //header的类型
	w.Header().Add("Access-Control-Expose-Headers", "upParam")
	if r.Method == "OPTIONS" {
		return
	}
	var videoSample definition.Videos
	var uploadBlock definition.UploadBlock

	//获取传输类型
	upParam := r.Header.Get("upParam")
	if upParam == "" {
		common.Log("upParam is nill")
		respond(w, 0, 0, definition.RetFailed, "", definition.Msg_TranType_Nill)
		return
	}
	common.Log1("upParam:", upParam)
	err := json.Unmarshal([]byte(upParam), &uploadBlock)
	if err != nil {
		common.Log("uploadBlock:" + err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	common.Log1("uploadblock:", uploadBlock)
	tranType := uploadBlock.TranType

	Offset, _ := strconv.ParseInt(uploadBlock.Offset, 10, 64)
	videoSample.Offset = Offset

	Filesize, _ := strconv.ParseInt(uploadBlock.FileSize, 10, 64)
	videoSample.FileSize = Filesize

	videoSample.Md5 = uploadBlock.FileMd5

	checkRes := models.CheckVideoType(uploadBlock.FileType)
	if checkRes == false {
		common.Log(" get filetype invalid")
		respond(w, 0, 0, definition.RetTypeInvalid, videoSample.Md5, definition.Msg_FileType_Nill)
		//responBlock.Retcode= definition.RetTypeInvalid
		// responBlock.Errmsg= definition.Msg_FileType_Nill
		// data6, _ := json.Marshal(responBlock)
		// responJson(w, data6)
		return
	}
	videoSample.FileType = uploadBlock.FileType
	// videoSample.FileName =uploadBlock.FileName

	m, _ := url.ParseQuery(uploadBlock.FileName)
	for key, _ := range m {
		videoSample.FileName = key
	}
	common.Log1("filname:", videoSample.FileName)
	// 检查chunksize是否等于ContentLength
	/*	if int64(chunksize) != r.ContentLength {
		common.Log(fmt.Sprintf("chunksize[%v] ContentLength[%v]\r\n", chunksize, r.ContentLength))
		handleInvalidParam(w, errors.New("chunksize not equal ContentLength"))
		return
	}*/

	common.Log(fmt.Sprintf("client:{ filetype[%v] filesize[%v] filemd5[%v] offset[%v] tranType[%v]}\r\n",
		videoSample.FileType, videoSample.FileSize, videoSample.Md5, videoSample.Offset, tranType))
	//判断传输类型
	// 如果是第一次传输
	if tranType == "cmd" {
		eoffset, res := models.NewVideo(videoSample)
		if res == 1 { //传输完成
			respond(w, eoffset, 0, definition.RetSucceed, videoSample.Md5, definition.Msg_Ok)
		} else {
			if res == 2 {
				respond(w, eoffset, blocksize, definition.RetFatal, "", definition.Msg_Internal_Error)
			}
			respond(w, eoffset, blocksize, definition.RetContinue, "", definition.Msg_Continue)
		}
		return
	}

	n, isCompleted, res, err := writeChunk(videoSample, r.Body)
	common.Log(fmt.Sprintf("write chank : next translate offse is %d, complete status is %d,result is %d", n, isCompleted, res))
	if res == false {
		common.Log("respons write fail")

		respond(w, 0, 0, definition.RetFatal, "", definition.Msg_Internal_Error)
		return
	}
	if err != nil { //要求客户端重传
		respond(w, Offset, blocksize, definition.RetFailed, "", definition.Msg_Resend)
		return
	}
	if isCompleted {
		respond(w, n, blocksize, definition.RetSucceed, videoSample.Md5, definition.Msg_Ok)
		return
	}
	respond(w, n, blocksize, definition.RetContinue, "", definition.Msg_Continue)
}

/*
func handleInvalidParam(w http.ResponseWriter, err error) bool {
	ErrInvalidParam := errors.New("invalid param")
	ErrStatusCodes := map[error]int{
		ErrInvalidParam: http.StatusBadRequest,
	}
	if err != nil {
		w.Header().Set("errmsg", ErrInvalidParam.Error())
		w.WriteHeader(ErrStatusCodes[ErrInvalidParam])
		return true
	}
	return false
}
*/
/* --------------------------------------------------------------------------*/
/**
* @brief respond           返回客户端
*
* @param http.ResponseWriter
* @param offset
* @param int64
* @param string
*
* @returns
 */
/* ----------------------------------------------------------------------------*/
func respond1(w http.ResponseWriter, offset int64, blockSize int64, retcode int64, md5 string, errmsg string) {
	var playDomain string
	playDomain = ""
	if retcode == definition.RetSucceed {
		domain := models.GetDomainForPlay(md5)
		if domain != "" {
			playDomain = domain
		}
	}

	w.Header().Set("playUrl", playDomain)
	w.Header().Set("offset", strconv.FormatInt(offset, 10))
	w.Header().Set("blockSize", strconv.FormatInt(blockSize, 10))
	w.Header().Set("retcode", strconv.FormatInt(retcode, 10))
	w.Header().Set("errmsg", errmsg)
}
func respond(w http.ResponseWriter, offset int64, blockSize int64, retcode int64, md5 string, errmsg string) {
	var playDomain string
	playDomain = ""
	if retcode == definition.RetSucceed {
		domain := models.GetDomainForPlay(md5)
		if domain != "" {
			playDomain = domain
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")  //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
	w.Header().Add("Access-Control-Expose-Headers", "*")
	w.Header().Set("content-type", "application/json")
	var responBlock definition.ResponBlock
	responBlock.Retcode = retcode
	responBlock.Errmsg = errmsg
	responBlock.Offset = offset
	responBlock.PlayUrl = playDomain
	responBlock.BlockSize = blockSize
	output, _ := json.Marshal(responBlock)
	w.Write(output)
	return
}
func verifyFileType(fileType string) bool {
	return true
}

/*
func writeCmdRespond(w http.ResponseWriter, blockSzie, retcode int64, errmsg string) {
	w.Header().Set("type", "cmdRest")
	w.Header().Set("blockSize", strconv.FormatInt(blockSzie, 10))
	w.Header().Set("retcode", strconv.FormatInt(retcode, 10))
	w.Header().Set("errmsg", errmsg)
}*/
