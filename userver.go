package main

import (
	"net/http"

	"./controlers"
	//"./models"
	//	"fmt"
	//"os/exec"
	// "bytes"
	//"io/ioutil"
	// "net/url"
)

func main() {
	/*	temp := "abcde/data/sbc.mp4"
		temp1 := common.ReplaceFileType(temp, "m3u9")
		fmt.Println("replace res:" + temp1)
	*/
	/*
	   	subtitle,color,size,startpost := models.GetSubtitle()
	   	fmt.Println("subtitle,color,size,startpost",subtitle,color,size,startpost)
	   	walterPlace:=models.GetWatermarkPlace()
	   	fmt.Println("walterplace:",walterPlace)
	       fragment := models.GetFragmentDuartion()
	   	fmt.Println("fragment:", fragment )
	   	var testStr string
	       testStr = "%e4%b8%ad%e5%9b%bd%e4%ba%ba"
	   	fmt.Println("testStr:", testStr)
	   	 m, _ := url.ParseQuery(testStr)
	       fmt.Println(m)
	       var name string
	   for key, value := range m {
	       fmt.Println("Key:", key, "Value:", value)
	       name = key
	       fmt.Println("name:",name)
	   }
	*/
	/*
		res := models.CheckVideoType("mp4")
		if res == true {
			fmt.Println("true")
		}else{
			fmt.Println("false")
		}
	*/
	/*
		fmt.Println("ffmpeg test")
		cmd := "ffmpeg -i test.mp4"
		execCmd(cmd)
	*/
	//fmt.Println("rest:",rest)
	go controlers.TranCodeRun()
	http.HandleFunc("/upload", controlers.HandleUpload)
	http.HandleFunc("/notify", controlers.CloudNotifyProcess)
	http.HandleFunc("/delVideo", controlers.DelVideoProcess)
	http.ListenAndServe(":8989", nil)
	//res :=controlers.ZimuHiProcess("test.mp4")
	//res :=controlers.M3u8Process("test.mp4")
	//res :=controlers.WaterVideo("test.mp4")
	//if res =="" {
	//fmt.Println("finish")
	//}
	//controlers.WaterVideo("test.mp4")
}
