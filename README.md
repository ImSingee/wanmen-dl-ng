# 某某培训机构视频下载器

## 使用指引

[https://coda.io/d/_dgfjTL5H1n1/_suy2E#_luFTb](https://coda.io/d/_dgfjTL5H1n1/_suy2E#_luFTb)

交流需求请入 QQ 群：689898623

> 本项目在 Linux / MacOS 下测试通过，尽力做了 Windows 支持但不保证可用性
> 
> 本项目依赖 ffmpeg，请查看文档安装


## 配置 Token 与下载路径

创建 config.json 文件，写入类似下面的内容（如不清楚怎么获取请查看上述文档）

```json
{
    "Authorization":"Bearer eTc2LCJpc3MiOiJ1cm46YXBpIn0.SUNiOJ7cM-ngFb7Yb9qSq5nAEiuUL2oQ5WWtr91_ONQ",
    "DownloadTo": "/path/to/dir"
}
```

- Authorization 的内容修改为电脑登录后 localstorage 中 authorization 的值
- DownloadTo 的内容修改为下载目标路径（Windows 下类似 `D:\\Downloads\\XXX`，注意需要两个 `\`）

## 下载

直接 `./wanmen-dl download <课程ID>`

课程 ID 来源为电脑端课程播放页面 `https://www.wanmen.org/courses/aaa/lectures/bbb` 中的 `aaa` 部分

如果提示 ID 错误，请执行 `./wanmen-dl register <课程ID>` 注册课程（仅出错时需要执行）

## 批量下载

创建 `to_download` 文件，每行一个课程 ID

运行 `./wanmen-dl download-all` 进行批量下载（可以使用 `-c` 控制并发，具体请查看 `./wanmen-dl download-all -h` 帮助文档

> 可以创建多个文件，对于任何文件名不为 to_download 的文件，请使用 `./wanmen-dl download-all file1` 明确传递文件名

> 这一批量 **不会** 并发下载多个课程，如果需要请创建多个 list 并运行多个 download-all 程序

## 检查下载状态

运行 `./wanmen-dl verify <course-id>` 检查下载完整性

运行 `./wanmen-dl verify-all` 检查列表中的课程下载完整性（参数同批量下载）

## 删除中间文件

某些异常情况会产生一些中间文件，可以使用 `./wanmen-dl clean <course-id>` 清除，同样这个命令也有针对批量的 `clean-all` 版本

## 后台下载

本程序采用单进程单线程的方式，并未针对后台下载做特殊优化。如需后台下载请使用 tmux + 下载脚本进行

## 免责声明

该脚本来源于网络，并非本人编写，本人从未对该培训机构和相关的视频加密提供方进行逆向工程，本人也从未使用过该脚本

该脚本的发布仅为学习使用，本人不承担使用者的行为所带来的任何法律后果

该脚本使用 [No License](https://choosealicense.com/no-permission/) 协议发布，版权人未知，因此任何人都不得对该项目进行修改、拷贝、分发、用于商业用途
