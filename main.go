package torpedo_common

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
	"gopkg.in/h2non/filetype.v1"
)

const User_Agent = "Mozilla/5.0 (https://github.com/tb0hdan/torpedo; tb0hdan@gmail.com) Go-http-client/1.1"

type Utils struct {
	logger *log.Logger
}

func (cu *Utils) SetLoggerPrefix(prefix string) (logger *log.Logger) {
	logger = cu.NewLog(prefix)
	cu.SetLogger(logger)
	return
}

func (cu *Utils) NewLog(prefix string) (logger *log.Logger) {
	logger = log.New(os.Stdout, fmt.Sprintf("%s: ", prefix), log.Lshortfile|log.LstdFlags)
	return
}

func (cu *Utils) SetLogger(logger *log.Logger) {
	cu.logger = logger
	return
}

func (cu *Utils) HTTPClient(req_type, url, content_type string, body *strings.Reader) (result []byte, err error) {
	client := &http.Client{}

	if body == nil {
		body = strings.NewReader("")
	}
	req, err := http.NewRequest(req_type, url, body)
	if err != nil {
		cu.logger.Fatalln(err)
	}

	req.Header.Set("User-Agent", User_Agent)
	if content_type != "" {
		req.Header.Set("Content-Type", content_type)
	}

	resp, err := client.Do(req)
	if err != nil {
		cu.logger.Fatalln(err)
	}

	defer resp.Body.Close()
	result, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		cu.logger.Fatalln(err)
	}
	return
}

func (cu *Utils) GetURLBytes(url string) (result []byte, err error) {
	result, err = cu.HTTPClient("GET", url, "", nil)
	return
}

func (cu *Utils) PostURLBytes(url, content_type string, body *strings.Reader) (result []byte, err error) {
	result, err = cu.HTTPClient("POST", url, content_type, body)
	return
}

func (cu *Utils) GetURLUnmarshal(url string, result interface{}) (err error) {
	data, err := cu.GetURLBytes(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, result)
	return
}

func (cu *Utils) PostURLUnmarshal(url, content_type string, body *strings.Reader, result interface{}) (err error) {
	data, err := cu.PostURLBytes(url, content_type, body)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, result)
	return
}

func (cu *Utils) PostURLFormUnmarshal(url string, data url.Values, result interface{}) (err error) {
	body := strings.NewReader(data.Encode())
	err = cu.PostURLUnmarshal(url, "application/x-www-form-urlencoded", body, result)
	return
}

func (cu *Utils) GetMIMEType(fname string) (mimetype, extension string, is_image bool, err error) {
	// Read a file
	buf, err := ioutil.ReadFile(fname)

	if err != nil {
		cu.logger.Printf("GetMIMEType could not read file %s", fname)
		return
	}

	// We only have to pass the file header = first 261 bytes
	head := buf[:261]

	kind, err := filetype.Match(head)
	if err != nil {
		cu.logger.Printf("Mimetype unkwown: %s", err)
		return
	}

	mimetype = kind.MIME.Value
	extension = kind.Extension
	is_image = filetype.IsImage(head)
	return
}

func (cu *Utils) DownloadToTmp(url string) (fname string, mimetype string, is_image bool, err error) {
	img, _ := cu.GetURLBytes(url)
	tmpfile, err := ioutil.TempFile("/tmp", "torpedo")
	if err != nil {
		cu.logger.Fatal(err)
	}

	if _, err := tmpfile.Write(img); err != nil {
		cu.logger.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		cu.logger.Fatal(err)
	}
	fname = tmpfile.Name()
	mimetype, _, is_image, err = cu.GetMIMEType(fname)
	return
}

func GetLimitPage(search_string string, limit_default, page_default int) (limit, page int) {
	limit = limit_default
	page = page_default
	r := regexp.MustCompile(`((limit|page)\:\d+)`)
	result := r.FindAllStringSubmatch(search_string, -1)
	for _, group := range result {
		split := strings.Split(group[0], ":")
		switch split[0] {
		case "limit":
			i, err := strconv.Atoi(split[1])
			if err == nil {
				limit = i
			}
		case "page":
			i, err := strconv.Atoi(split[1])
			if err == nil {
				page = i
			}
		default:
			fmt.Println("Unknown!")
		}
	}
	return
}

func GetRequestedFeature(full_command string, usage ...string) (requestedFeature, command, message string) {
	// Support multiple commands within single function
	requestedFeature = strings.Split(full_command, " ")[0]
	command = strings.TrimSpace(strings.TrimLeft(full_command, requestedFeature))
	if len(usage) == 0 {
		message = fmt.Sprintf("Usage: %s string\n", requestedFeature)
	} else {
		message = fmt.Sprintf("Usage: %s %s\n", requestedFeature, usage[0])
	}
	return
}

func ChannelsUploadImage(channels []string, fname, fpath, ftype string, api_i interface{}) {
	/*
			channels := []string{channel.(string)}
		filename := fmt.Sprintf("%s.png", command)
		common.ChannelsUploadImage(channels, filename, filepath, mimetype, api) */
	parameters := slack.FileUploadParameters{File: fpath, Filetype: ftype,
		Filename: fname, Title: fname,
		Channels: channels}
	api := api_i.(slack.Client)
	api.UploadFile(parameters)
}

func UnformatURL(url string) (newurl string) {
	re := regexp.MustCompile("[<>]")
	newurl = strings.TrimSpace(re.ReplaceAllString(url, ""))
	return
}

func FileExists(fpath string) (exists bool) {
	// TODO: Find a way around this, os.IsExist expects an error and we don't have one yet
	exists = true
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		exists = false
	}
	return
}

func MD5Hash(message string) (result string) {
	my_hash := md5.New()
	io.WriteString(my_hash, message)
	result = fmt.Sprintf("%x", my_hash.Sum(nil))
	return
}

func SHA1Hash(message string) (result string) {
	my_hash := sha1.New()
	io.WriteString(my_hash, message)
	result = fmt.Sprintf("%x", my_hash.Sum(nil))
	return
}

func SHA256Hash(message string) (result string) {
	my_hash := sha256.New()
	io.WriteString(my_hash, message)
	result = fmt.Sprintf("%x", my_hash.Sum(nil))
	return
}

func SHA512Hash(message string) (result string) {
	my_hash := sha512.New()
	io.WriteString(my_hash, message)
	result = fmt.Sprintf("%x", my_hash.Sum(nil))
	return
}

func GetStripEnv(envvar string) (result string) {
	result = os.Getenv(envvar)
	result = strings.TrimLeft(result, "'")
	result = strings.TrimRight(result, "'")
	return
}
