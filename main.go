package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var fileName string
var base string
var urlAll []string
var url string

//初始化参数
func init() {
	flag.StringVar(&url, "u", "https://archive.apache.org/dist/tomcat/tomcat-3/bin/netware/", "url")
	flag.StringVar(&fileName, "f", "result/text.txt", "file name")
	flag.Parse()
	base = getBase(url)
}

//getBase 获取url的主域
func getBase(url string) string {
	r, err := regexp.Compile("http[s]*://[a-zA-Z_.-]+")
	if err != nil {
		log.Printf("err: %v\n", err)
	}
	str := r.Find([]byte(url))
	return string(str)
}

//isFile，判断url指向的是否是文件
func isFile(url string) bool {
	url = strings.Replace(url, base, "", -1)
	return strings.Contains(url, ".")
}

//保存
func save(url string) {
	fmt.Printf(url + "\n")
	urlAll = append(urlAll, url)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer file.Close()
	write := bufio.NewWriter(file)

	n, err := write.WriteString(url + " \r\n")
	if n == 0 || err != nil {
		fmt.Printf("err: %v\n", err)
	}
	//Flush将缓存的文件真正写入到文件中
	write.Flush()

}

//传入一个url，判断是否有重复的，算法有待改进
func isContain(url string) bool {
	for i := 0; i < len(urlAll); i++ {
		if urlAll[i] == url {
			return true
		}
	}
	return false
}

//adsUrl 传入获取的url，发送请求的urlHttp，如果包含在urlHttp，返回一个url的绝对路径，否则返回false。
//避免跳转的其他非漏洞链接，避免向上层跳转
func adsUrl(url, urlHttp string) (string, bool) {
	if url[0] == '/' {
		if len(url) > 1 {
			if url[1] != '/' {
				if !strings.Contains(base+url, urlHttp) {
					return "", false
				}
				return base + url, true
			}
			return "", false
		}
		if len(url) == 1 {
			if !strings.Contains(base+url, urlHttp) {
				return "", false
			}
			return base + url, true
		}
	}
	return urlHttp + "/" + url, true
}

//getUrls 传入一个url，返回里面的 href 内容列表
func getUrls(url string) []string {
	var resp *http.Response
	var errHttp error
	for i := 0; i < 3; i++ {
		resp, errHttp = http.Get(url)
		if errHttp != nil {
			continue
		} else {
			break
		}
	}
	if errHttp != nil {
		fmt.Printf("err: %v\n", errHttp)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	reg, err := regexp.Compile("href\\s*=\\s*\"([a-zA-Z0-9/_.-]+)\"") //这里的正则可以当参数
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil
	}
	result := reg.FindAllStringSubmatch(string(body), -1)
	var list []string
	for _, text := range result {
		u, flg := adsUrl(text[1], url)
		if !flg {
			continue
		}
		list = append(list, u)
	}
	return list
}

// recursion 传入一个url列表，递归寻找里面包含的所有url
func recursion(urls []string) {
	//递归结束判断
	if len(urls) == 0 {
		return
	}
	var list []string
	for i := range urls {
		if !isFile(urls[i]) && !isContain(urls[i]) {
			list = append(list, getUrls(urls[i])...)
		}
		save(urls[i])
	}
	recursion(list)

}

func main() {

	var urls []string
	urls = append(urls, url)
	recursion(urls)
}
