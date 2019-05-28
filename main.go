package main

import (
    "bytes"
    "crypto/tls"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "net/http"
    "net/smtp"
    "regexp"
    "strconv"
    "strings"
    "time"
)
//smtp的服务器这是QQ邮箱的
const host=`smtp.qq.com`
//smtp端口只用了加密端口
const port=465
//你的邮件地址这个当然是乱填的
const from="2117143053@qq.com"
//smtp服务商提供给你的密码
const password="asfsafsaffdsdf"
//标题
const title="求职信"
//信件内容模板支持一定格式的html
const templateStr=`<h1>求职信</h1>
<h2>你好！{{.creator.name}}</h2>求职：{{.title}}
`
var auth=smtp.PlainAuth("",from,password,host)
var tpl=template.Must(template.New("messageTpl").Parse(templateStr))








type  Mail struct{
    auth smtp.Auth
    head map[string]string
    message string
    attachments []Attachment
}
type Attachment struct{
    id string
    name string
    file string
    contentType string


}

func getTimeParam(time time.Time)string{
  t:=strconv.FormatInt(time.UnixNano(),10)
   return t[0:13]


}
func getBilibili(time time.Time,limit int,offset int) map[string]interface{}{
    client:=&http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
    }
    timeString:=getTimeParam(time)
    //bilibili社招页面地址当前为上海，技术类，其他类型请自行更改参数，非技术岗位请在技术岗陪同下使用此代码
    //参数请在技术陪同下在这个页面https://www.bilibili.com/blackboard/join-list.html#/detail/7826c224-4054-4bce-802b-459f725b19a9找
    url:=`https://api.bilibili.com/x/web-goblin/recruit?mode=social&route=v1/jobs&limit=%s&offset=%s&locationIds=[11052]&zhinengId=15931&time=%s`
    url=fmt.Sprintf(url,limit,offset,timeString)
    resp,err:=client.Get(url)
    defer resp.Body.Close()
    if err!=nil{

       panic(err)
    }
    result,err2:=ioutil.ReadAll(resp.Body)
    if err2!=nil{
      panic(err2)

    }
    ret:=make(map[string]interface{})
    json.Unmarshal(result,&ret)
    return ret["data"].(map[string]interface{})







}
func match(job map[string]interface{})bool{
    var reg1,reg2,reg3,reg4,reg5 *regexp.Regexp
    var err error
    reg1,err=regexp.Compile(".*(?i:JAVA).*")
    reg2,err=regexp.Compile(".后端.*")
    reg3,err=regexp.Compile(".*(?i:Android).*")
    reg4,err=regexp.Compile(".*(?i:Linux).*")
    reg5,err=regexp.Compile(".*(?i:go).*")

    if err!=nil {
        panic(err)
    }
    return reg1.Match([]byte(job["title"].(string)))||
        reg2.Match([]byte(job["title"].(string)))||
        reg3.Match([]byte(job["title"].(string)))||
        reg4.Match([]byte(job["title"].(string)))||
        reg5.Match([]byte(job["title"].(string)))||
        reg1.Match([]byte(job["description"].(string)))||
        reg2.Match([]byte(job["description"].(string)))||
        reg3.Match([]byte(job["description"].(string)))||
        reg4.Match([]byte(job["description"].(string)))||
        reg5.Match([]byte(job["description"].(string)))


}

//https://api.bilibili.com/x/web-goblin/recruit?mode=social&route=v1/jobs&limit=14&offset=70&locationIds=[11052]&zhinengId=15931&time=1558849837634
//https://api.bilibili.com/x/web-goblin/recruit?mode=social&route=v1/jobs&limit=14&offset=70&locationIds=[11052]&zhinengId=15931&time=1558849837634

func (mail *Mail) Create(from string,cc,bcc,subject,message string,attachments []Attachment){
    head:=make(map[string]string)
    head["From"]=from
    head["Cc"]=cc
    head["Bcc"]=bcc
    head["Subject"]=subject
    mail.auth=auth
    mail.head=head
   mail.message=message
   mail.attachments=attachments
}
func (mail *Mail) Send(to string){
    mail.head["To"]=to
   var err error
    body:=bytes.NewBuffer(nil)
    keys:=[]string{"From","To","Cc","Bcc","Subject"}
    for _,v:=range keys{
        body.WriteString(fmt.Sprintf("%s:%s\r\n",v,mail.head[v]))
    }
    boundary:="jiayang"+strconv.FormatInt(time.Now().UnixNano(),10)
    body.WriteString(fmt.Sprintf("Content-Type:multipart/related;boundary=%s\r\n",boundary))
    body.WriteString(fmt.Sprintf("Date:%s\r\n",time.Now().String()))
    body.WriteString("\r\n")
    body.WriteString(fmt.Sprintf("\r\n--%s\r\n",boundary))
    body.WriteString("Content-Type: text/html;charset=UTF-8 \r\n\r\n")
    body.WriteString(mail.message)
    body.WriteString("\r\n")
    for _,attachment:=range mail.attachments{
        file,err:=ioutil.ReadFile(attachment.file)
        if err!=nil{
            panic(err)
        }

        base64temp:=make([]byte,base64.StdEncoding.EncodedLen(len(file)))
        base64.StdEncoding.Encode(base64temp,file)
        body.WriteString(fmt.Sprintf("\r\n--%s\r\n",boundary))
        body.WriteString("Content-Transfer-Encoding:base64\r\n")
        body.WriteString(fmt.Sprintf("Content-Type:%s;name=%s\r\n",attachment.contentType,attachment.name))
        body.WriteString(fmt.Sprintf("Content-ID:%s\r\n\r\n",attachment.id))

        for index,value:=range base64temp{
            body.WriteByte(value)
            if index+1%76==0 {
                body.WriteString("\r\n")
            }


        }
        body.WriteString("\r\n")




    }

    body.WriteString(fmt.Sprintf("\r\n--%s--",boundary))
    fmt.Println(body.String())
     fmt.Println(mail)

    conn,err:=tls.Dial("tcp",fmt.Sprintf("%s:%s",host,strconv.Itoa(port)),nil)
    if err!=nil{
        panic(err)
    }
    client,err:=smtp.NewClient(conn,host)
    if err!=nil{
        panic(err)
    }
    if ok, _ := client.Extension("AUTH"); ok{
        if err=client.Auth(mail.auth);err!=nil{
            panic(err)
        }
    }



    err=client.Mail(from)
    if err!=nil{
        panic(err)
    }
    for _,v:=range strings.Split(mail.head["To"],";"){
       err= client.Rcpt(v)
        if err!=nil{
            panic(err)
        }

    }
   writer,err:=client.Data()
   if err!=nil{
       panic(err)
   }
   defer writer.Close()
   writer.Write(body.Bytes())









}



func main(){
    defer func() {
        if err:=recover();err!=nil{
            fmt.Print(err)
        }

    }()

    result:=getBilibili(time.Now(),14,0)
    result=getBilibili(time.Now(),int(result["total"].(float64)),0)
    jobs:=result["jobs"].([]interface{})
    count:=0
    for index,job:=range jobs{
        if match(job.(map[string]interface{})) {
            count++
            message:=bytes.NewBuffer(nil)
            tplerr:=tpl.Execute(message,jobs[index])
            if tplerr!=nil{
                fmt.Print(tplerr)
            } else{
                fmt.Print(message.String())

                    mail:=&Mail{}
                    //想要传送附件在这里contetnType请参考mime
                    //其他说实话好像没有什么用
                    attachments:=[]Attachment{
                        {file:"G:/1.txt",id:"1.txt",contentType:"text/plain",name:"1.txt"},
                        {file:"G:/2.txt",id:"testzip",contentType:"application/zip",name:"test.zip"},
                    }
                    mail.Create(from,"","",title,message.String(),attachments)
                    mail.Send(jobs[index].(map[string]interface{})["creator"].(map[string]interface{})["email"].(string))
                    fmt.Printf("\n发邮件至%s\n",jobs[index].(map[string]interface{})["creator"].(map[string]interface{})["email"].(string) )
                    //睡眠时间避免你的邮箱频率不够无奈啊没有RMB是硬伤
                    time.Sleep(10*time.Second)
            }

            }
        fmt.Printf("*****************************************\n")



    }

    }
