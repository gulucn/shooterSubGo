# shooterSubGo
shooter.cn subtitle downloader with Golang

## 编译脚本

```bash
E:\thirdparty\shooterSubGo>go build
E:\thirdparty\shooterSubGo>shooterSubGo.exe -h
Usage:
    shooterSubGo.exe [option] path
    -d    need output debug info
    -ext string
        check ext file list,default include:avi/mp4/mkv/rm/rmvb
    -h    print help info
```

## 使用
### 指定文件
  
```bash
E:\thirdparty\shooterSubGo>shooterSubGo.exe  F:\TDDownload\runningman2010\E001.1
00711.時代廣場.avi
[INFO] handle name: F:\TDDownload\runningman2010\E001.100711.時代廣場.avi
[INFO] download file F:\TDDownload\runningman2010\E001.100711.時代廣場.chn.ass o
k,spend:3.7182127s
[INFO] write file F:\TDDownload\runningman2010\E001.100711.時代廣場.chn1.ass.del
ay ok,delay:100
[INFO] download file F:\TDDownload\runningman2010\E001.100711.時代廣場.chn1.ass
ok,spend:5.0422884s
[INFO] write file F:\TDDownload\runningman2010\E001.100711.時代廣場.chn2.ass.del
ay ok,delay:300
[INFO] download file F:\TDDownload\runningman2010\E001.100711.時代廣場.chn2.ass
ok,spend:5.0372881s
E:\thirdparty\shooterSubGo>  
``` 

### 指定目录
```bash
E:\thirdparty\shooterSubGo>shooterSubGo.exe  F:\TDDownload\runningman2010\
```
