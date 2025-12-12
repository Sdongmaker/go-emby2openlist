<div align="center">
  <img height="150px" src="./assets/logo.png"></img>
</div>

<h1 align="center">go-emby2openlist</h1>

<div align="center">
  <a href="https://github.com/AmbitiousJun/go-emby2openlist/tree/v2.3.2"><img src="https://img.shields.io/github/v/tag/AmbitiousJun/go-emby2openlist"></img></a>
  <a href="https://hub.docker.com/r/ambitiousjun/go-emby2openlist/tags"><img src="https://img.shields.io/docker/image-size/ambitiousjun/go-emby2openlist/v2.3.2"></img></a>
  <a href="https://hub.docker.com/r/ambitiousjun/go-emby2openlist/tags"><img src="https://img.shields.io/docker/pulls/ambitiousjun/go-emby2openlist"></img></a>
  <a href="https://github.com/AmbitiousJun/go-emby2openlist/releases/latest"><img src="https://img.shields.io/github/downloads/AmbitiousJun/go-emby2openlist/total"></img></a>
  <a href="https://goreportcard.com/report/github.com/AmbitiousJun/go-emby2openlist/v2"><img src="https://goreportcard.com/badge/github.com/AmbitiousJun/go-emby2openlist/v2"></img></a>
  <img src="https://img.shields.io/github/stars/AmbitiousJun/go-emby2openlist"></img>
  <img src="https://img.shields.io/github/license/AmbitiousJun/go-emby2openlist"></img>
</div>

<div align="center">
  Go 语言编写的 Emby + OpenList 网盘直链反向代理服务，深度适配阿里云盘转码播放。
</div>

## 小白必看

**网盘直链反向代理**:

正常情况下，Emby 通过磁盘挂载的形式间接读取网盘资源，走的是服务器代理模式，看一个视频时数据链路是：

> 客户端 => Emby 源服务器 => 磁盘挂载服务 => OpenList => 网盘
>
> 客户端 <= Emby 源服务器 <= 磁盘挂载服务（将视频数据加载到本地，再给 Emby 读取） <= OpenList <= 网盘

这种情况有以下局限：

1. 视频经过服务器中转，你看视频的最大加载速度就是服务器的上传带宽
2. 如果服务器性能不行，能流畅播放 1080p 就谢天谢地了，更别说 4K
3. ...

使用网盘直链反向代理后，数据链路：

> 客户端 => Emby 反代服务器 => Emby 源服务器 （请求 Emby Api 接口）
>
> 客户端 <= Emby 反代服务器 <= Emby 源服务器 （返回数据）

对于普通的 Api 接口，反代服务器将请求反代到源服务器，再将合适的结果进行缓存，返回给客户端

对于客户端来说，这一步和直连源服务器看不出差别

> 客户端 => Emby 反代服务器 => OpenList => 网盘 （请求视频直链）
>
> 客户端 <= Emby 反代服务器 <= OpenList <= 网盘 （返回视频直链，并给出重定向响应）
>
> 客户端 => 网盘（客户端拿着网盘的直链直接观看，此时已经没有服务器的事情了，故不会再消耗服务器流量）

这种方式的好处：

1. 观看时加载速度拉满（前提是有网盘会员）
2. 在客户端处解码，能不能看 4K 取决于你电视盒子的性能

## 使用前须知

1. 本项目初衷: 易用轻巧、小白友好、深度适配阿里云盘 (如果你使用本项目观看其他网盘时出现问题，也欢迎到 issue 区反馈，我会尽量适配它)

2. 如果你有更复杂的需求, 推荐使用功能更完善的反向代理服务：[bpking1/embyExternalUrl](https://github.com/bpking1/embyExternalUrl)

## 功能

- OpenList 网盘原画直链播放

- Strm 直链播放

- [OpenList 本地目录树生成](#使用说明-openlist-本地目录树生成)

- [自定义注入 js/css（web）](#使用说明-自定义注入-web-js-脚本)

- OpenList 网盘转码直链播放（阿里云盘）

- **🆕 [对象存储(OSS) 302重定向 + CDN Type-A 鉴权](#使用说明-对象存储oss-重定向)（魔改功能）**

  > **适用场景**：媒体文件已存储在公共对象存储（S3/OSS/COS等），需要带CDN鉴权的直链播放
  >
  > **主要特性**：
  > - 直接重定向到对象存储，无需通过OpenList中转
  > - 支持CDN Type-A鉴权算法，自动生成带时效的auth_key
  > - 响应头携带API Key用于源站访问验证
  > - 自动处理中文路径URL编码
  > - 可与OpenList模式共存，通过配置切换

  > **是否消耗三方流量包流量**：🙅
  >
  > **非会员是否限速**：自行测试
  >
  > 
  >
  > 示例图 ↓：
  >
  > <img src="assets/2024-08-31-17-15-53.jpg" />
  >
  > 
  >
  > 转码资源直链已达到可正常使用的标准，Emby Web, Emby for AndroidTV 以及其他大部分客户端都可以正常播放，并且不会因为直链过期而中断
  >
  > 
  >
  > 局限：
  >
  > 如果是有多个内置音频的，转码直链只能播放其中的默认音频
  >
  > 视频本身的内封字幕会丢失，不过若存在转码字幕，也会适配到转码版本的 PlaybackInfo 信息中，示例图 ↓：
  >
  > <img src="assets/2025-04-27-20-15-06.png"/>
  
- websocket 代理

- 客户端防转码（转容器）

- 缓存中间件，实际使用体验不会比直连源服务器差

- 字幕缓存（字幕缓存时间固定 30 天）

  > 目前还无法阻止 Emby 去本地挂载文件上读取字幕
  >
  > 带字幕的视频首次播放时，Emby 会调用 FFmpeg 将字幕从本地文件中提取出来，再进行缓存
  >
  > 也就是说：
  >
  > - 首次提取时，速度会很慢，有可能得等个大半天才能看到字幕（使用第三方播放器【如 `MX player`, `Fileball`】可以解决）
  > - 带字幕的视频首次播放时，还是会消耗服务器的流量

- 直链缓存（为了兼容阿里云盘，直链缓存时间目前固定为 10 分钟，其他云盘暂无测试）

- 大接口缓存（OpenList 转码资源是通过代理并修改 PlaybackInfo 接口实现，请求比较耗时，每次大约 2~3 秒左右，目前已经利用 Go 语言的并发优势，尽力地将接口处理逻辑异步化，快的话 1 秒即可请求完成，该接口的缓存时间目前固定为 12 小时，后续如果出现异常再作调整）



## 已测试并支持的客户端

| 名称                                             | 最后测试版本 | 原画 | 其他说明（原画）                                             | 阿里转码 | 其他说明（阿里转码）                                         |
| ------------------------------------------------ | ------------ | ---- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| [`Gemby`](https://github.com/AmbitiousJun/gemby) | `v2.0.0`     | ✅    | ——                                                           | ✅        | ——                                                           |
| `Emby Web`                                       | `4.8.8.0`    | ✅    | ——                                                           | ✅        | 1. 转码字幕有概率挂载不上<br />2. 可以挂载原画字幕           |
| `Emby for iOS`                                   | ——           | ❓    | ~~没高级订阅测不了~~                                         | ❓        | ~~没高级订阅测不了~~                                         |
| `Emby for macOS`                                 | ——           | ❓    | ~~没高级订阅测不了~~                                         | ❓        | ~~没高级订阅测不了~~                                         |
| `Emby for Android`                               | `3.4.23`     | ✅    | ——                                                           | ✅        | ——                                                           |
| `Emby for AndroidTV`                             | `2.0.95g`    | ✅    | 遥控器调进度可能会触发直链服务器的频繁请求限制，导致视频短暂不可播情况 | ✅        | 无法挂载字幕                                                 |
| `Fileball`                                       | ——           | ✅    | ——                                                           | ✅        | ——                                                           |
| `Infuse`                                         | ——           | ✅    | 在设置中将缓存方式设置为`不缓存`可有效防止触发频繁请求       | ❌        | ——                                                           |
| `VidHub`                                         | ——           | ✅    | 仅测试至 `1.0.7` 版本                                        | ✅        | 仅测试至 `1.0.7` 版本                                        |
| `Stream Music`                                   | `1.3.8`      | ✅    | ——                                                           | ——       | ——                                                           |
| `Emby for Kodi Next Gen`                         | `11.1.13`    | ✅    | ——                                                           | ✅        | 1. 需要开启插件设置：**播放/视频转码/prores**<br />2. 播放时若未显示转码版本选择，需重置本地数据库重新全量扫描资料库<br />3. 某个版本播放失败需要切换版本时，必须重启 kodi 才能重新选择版本<br />4. 无法挂载字幕 |



## 前置环境准备

1. 已有自己的 Emby、OpenList 服务器

2. Emby 的媒体库路径（本地磁盘路径）是和 OpenList 挂载路径能够对应上的

   > 这一步前缀对应不上没关系，可以在配置中配置前缀映射 `path.emby2openlist` 解决

3. 需要有一个中间服务，将网盘的文件数据挂载到系统本地磁盘上，才能被 Emby 读取到（也可借助本项目的 [OpenList 目录树生成功能](#使用说明-openlist-本地目录树生成)）

   > 目前我知道的比较好用的服务有两个：[rclone](https://rclone.org/) 和 [CloudDrive2](https://www.clouddrive2.com/)(简称 cd2)
   >
   > 
   >
   > 如果你的网盘跟我一样是阿里云盘，推荐使用 cd2 直接连接阿里云盘，然后根路径和 OpenList 保持即可
   >
   > 在 cd2 中，找到一个 `最大缓存大小` 的配置，推荐将其设为一个极小值（我是 1MB），这样在刮削的时候就不会消耗太多三方权益包的流量
   >
   > 
   >
   > ⚠️ 不推荐中间服务直接去连接 OpenList 的 WebDav 服务，如果 OpenList Token 刷新失败或者是请求频繁被暂时屏蔽，会导致系统本地的挂载路径丢失，Emby 就会认为资源被删除了，然后元数据就丢了，再重新挂载回来后就需要重新刮削了。

4. 服务器有安装 Docker

5. Git

   > 非必须，如果你想体验测试版，就需要通过 Git 拉取远程源码构建
   >
   > 正式版可以直接使用现成的 Docker 镜像

## 使用 DockerCompose 部署安装

### 通过源码构建

1. 获取代码

```shell
git clone --branch v2.3.2 --depth 1 https://github.tbedu.top/https://github.com/AmbitiousJun/go-emby2openlist
cd go-emby2openlist
```

2. 拷贝配置

> 示例配置为完整版配置，首次部署可以参照[核心配置](https://github.com/AmbitiousJun/go-emby2openlist/issues/108#issuecomment-2928599051)优先跑通程序，再按需补充其他配置

```shell
cp config-example.yml config.yml
```

3. 根据自己的服务器配置好 `config.yml` 文件

关于路径映射的配置示例图：

![路径映射示例](assets/2024-09-05-17-20-23.png)

4. 编译并运行容器

```shell
docker-compose up -d --build
```

5. 浏览器访问服务器 ip + 端口 `8095`，开始使用

   > 如需要自定义端口，在第四步编译之前，修改 `docker-compose.yml` 文件中的 `8095:8095` 为 `[自定义端口]:8095` 即可

6. 日志查看

```shell
docker logs -f go-emby2openlist -n 1000
```

7. 修改配置的时候需要重新启动容器

```shell
# 修改 config.yml ...
docker-compose restart
```

8. 版本更新

```shell
# 获取到最新代码后, 可以检查一下 config-example.yml 是否有新增配置
# 及时同步自己的 config.yml 才能用上新功能

# 更新到正式版
docker-compose down
git fetch --tag
git checkout <版本号>
git pull
docker-compose up -d --build

# 更新到测试版 (仅尝鲜, 不稳定)
docker-compose down
git checkout main
git pull origin main
docker-compose up -d --build
```

9. 清除过时的 Docker 镜像

```shell
docker image prune -f
```

### 使用现有镜像

1. 准备配置

> 示例配置为完整版配置，首次部署可以参照[核心配置](https://github.com/AmbitiousJun/go-emby2openlist/issues/108#issuecomment-2928599051)优先跑通程序，再按需补充其他配置

参考[示例配置](https://github.com/AmbitiousJun/go-emby2openlist/blob/v2.3.2/config-example.yml)，配置好自己的服务器信息，保存并命名为 `config.yml`

2. 创建 docker-compose 文件

在配置相同目录下，创建 `docker-compose.yml` 粘贴以下代码：

```yaml
version: "3.1"
services:
  go-emby2openlist:
    image: ambitiousjun/go-emby2openlist:v2.3.2
    environment:
      - TZ=Asia/Shanghai
      - GIN_MODE=release
    container_name: go-emby2openlist
    restart: always
    volumes:
      - ./config.yml:/app/config.yml
      - ./ssl:/app/ssl
      - ./custom-js:/app/custom-js
      - ./custom-css:/app/custom-css
      - ./lib:/app/lib
      - ./openlist-local-tree:/app/openlist-local-tree
    ports:
      - 8095:8095 # http
      - 8094:8094 # https
```

3. 运行容器

```shell
docker-compose up -d --build
```

## 使用说明 ssl

**使用方式：**

1. 将证书和私钥放到程序根目录下的 `ssl` 目录中
2. 再将两个文件的文件名分别配置到 `config.yml` 中

**特别说明：**

在容器内部，已经将 https 端口写死为 `8094`，将 http 端口写死为 `8095`

如果需要自定义端口，仍然是在 `docker-compose.yml` 中将宿主机的端口映射到这两个端口上即可

**已知问题：**

可能有部分客户端会出现首次用 https 成功连上了，下次再打开客户端时，就自动变回到 http 连接，目前不太清楚具体的原因

## 使用说明 自定义注入 web js 脚本

**使用方式：** 将自定义脚本文件以 `.js` 后缀命名放到程序根目录下的 `custom-js` 目录后重启服务自动生效

**远程脚本：** 将远程脚本的 http 访问地址写入以 `.js` 后缀命名的文件后（**如编辑器报错请无视**）放到程序根目录下的 `custom-js` 目录后重启服务自动生效

**注意事项：** 确保多个不同的文件必须都是相同的编码格式（推荐 UTF-8）

**示例脚本：**

| 描述                  | 获取脚本                                                     | 自用优化版本                                                 |
| --------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 生成外部播放器按钮    | [ExternalPlayers.js](https://emby-external-url.7o7o.cc/embyWebAddExternalUrl/embyLaunchPotplayer.js) | ---                                                          |
| 首页轮播图            | [emby-swiper.js](https://raw.githubusercontent.com/newday-life/emby-web-mod/refs/heads/main/emby-swiper/emby-swiper.js) | [媒体库合并 + 每日清空缓存](https://github.com/AmbitiousJun/emby-css-js/raw/refs/heads/main/custom-js/emby-swiper.js) |
| 隐藏无图片演员        | [actorPlus.js](https://raw.githubusercontent.com/newday-life/emby-web-mod/refs/heads/main/actorPlus/actorPlus.js) | ---                                                          |
| 键盘 w/s 控制播放音量 | [audio-keyboard.js](https://github.com/AmbitiousJun/emby-css-js/blob/main/custom-js/audio-keyboard.js) | ---                                                          |

## 使用说明 自定义注入 web css 样式表

**使用方式：** 将自定义样式表文件以 `.css` 后缀命名放到程序根目录下的 `custom-css` 目录后重启服务自动生效

**远程样式表：** 将远程样式表的 http 访问地址写入以 `.css` 后缀命名的文件后（**如编辑器报错请无视**）放到程序根目录下的 `custom-css` 目录后重启服务自动生效

**注意事项：** 确保多个不同的文件必须都是相同的编码格式（推荐 UTF-8）

**示例样式：**

| 描述                 | 获取样式                                                    | 自用优化版本                                                 |
| -------------------- | ----------------------------------------------------------- | ------------------------------------------------------------ |
| 调整音量调整控件位置 | [音量条+控件修改.css](https://t.me/Emby_smzase1/74)         | ---                                                          |
| 节目界面样式美化     | [节目界面.txt](https://t.me/embycustomcssjs/10?comment=159) | [下拉框元素对齐](https://github.com/AmbitiousJun/emby-css-js/raw/refs/heads/main/custom-css/show-display.css) |

## 使用说明 OpenList 本地目录树生成

监控扫描 OpenList 目录树变更，在本地磁盘中生成并维护相应结构的目录树，可供 Emby 服务器直接扫描入库，并配合本项目进行直链反代，支持传统 Strm 文件以及附带元数据的虚拟文件生成。

> ⚠️ **提示：**
>
> 程序利用 Go 的并发优势，加快了扫描 OpenList 的速度，同时对于 ffmpeg 提取远程文件元数据操作进行了严格的并发控制，同一时刻只会有至多一个文件被提取元数据，扫库的风控风险虽大幅降低，但仍存在！

### 使用步骤

1. 升级到 go-emby2openlist `v2.2.0` 以上版本
2. 修改配置，按照自己的需求配置好 `openlist.local-tree-gen` 属性
3. 修改 `docker-compose.yml` 文件，将容器目录 `/app/openlist-local-tree` 以及 `/app/lib` 映射到宿主机中
4. 运行程序 开始自动扫描生成目录树
5. 将宿主机的目录树路径，映射到 Emby 容器中，即可扫描入库

### 不同使用场景的配置方式

由于 `openlist.local-tree-gen` 属性中的不同配置可能会有相互作用的情况，因此本说明按照具体的使用场景逐步递进地说明配置方式

1. 传统 Strm

   将需要转换成 strm 文件的文件容器配置到 `openlist.local-tree-gen.strm-containers` 属性中，以逗号分隔，不区分大小写，即可生效，示例配置：

   ```yaml
   openlist:
     local-tree-gen:
       enable: true
       strm-containers: mp4,mkv,mp3,flac
   ```

   **优点**：无需调用 ffmpeg，扫描速度极快，Emby 源端口 8096 也可能可以正常播放

   **缺点**：每个视频都需要使用 Emby 源端口至少播放一次后才能正常保存播放记录，配合反代服务体验不佳（`v2.2.12` 版本之后通过反代播放 Strm 也可正常记录播放进度）；无法获取视频的阿里转码版本

2. 虚拟文件

   此方式会在本地磁盘生成与远程文件**同名**的**空文件**（每个虚拟文件大小约 300B），对 Emby 来说就是大小为 0MB 的普通媒体文件。同理 Strm，将目标文件容器配置到 `openlist.local-tree-gen.virtual-containers` 属性中，以逗号分隔，不区分大小写，即可生效，示例配置：

   ```yaml
   openlist:
     local-tree-gen:
       enable: true
       virtual-containers: mp4,mkv
   ```

   下面介绍虚拟文件的两种处理方式，各有优缺，自行斟酌选用：

   默认情况下，程序会为每个媒体文件设置统一的时长元数据（3 小时，基本覆盖 99% 的媒体时长）。在实际使用时，Emby 能够正常记录播放进度，只不过在 UI 展示层面可能有点奇怪，如：播放时进度条完全不动（视频实际时长远远比 3 小时短）

   **优点**：无需调用 ffmpeg，扫描速度极快，风控风险低；Emby 能够正常记录播放进度

   **缺点**：视频时长固定写死为 3 小时，体验不佳；Emby 源端口 8096 无法播放

   可以通过开启 ffmpeg，来解析远程媒体的真实时长，并写入虚拟文件中来解决上述问题：

   ```yaml
   openlist:
     local-tree-gen:
       enable: true
       ffmpeg-enable: true
       virtual-containers: mp4,mkv
   ```

   使用这种方式后，程序会在扫描文件时，自动调用 ffmpeg 提取视频的真实时长，并写入本地虚拟文件中

   **优点**：视频时长真实，Emby 能够正常记录播放进度，且体验良好

   **缺点**：调用 ffmpeg 解析远程参数，有一定的风控风险；扫描速度相较第一种方式略慢一些；Emby 源端口 8096 无法播放

3. 音乐虚拟文件

   此方式类似于方式 2，扫描时除了提取文件的真实时长之外，还会提取音乐内嵌的标签元数据，写入到本地虚拟文件中（每个文件大小约 300KB~1MB 不等），使得 Emby 能够正常扫描解析音乐标签（标题、艺术家、海报、歌词等）以及音频真实播放时长。

   > ⚠️ 这种方式必须开启 ffmpeg 才能生效，且风控风险是 3 种方式中最高的一个，谨慎使用！
   >
   > 如果使用了此方式，但没有开启 ffmpeg，默认生成的是传统 Strm 文件

   将目标文件容器配置到 `openlist.local-tree-gen.music-containers` 属性中，以逗号分隔，不区分大小写，即可生效，示例配置：

   ```yaml
   openlist:
     local-tree-gen:
       enable: true
       ffmpeg-enable: true
       music-containers: mp3,flac
   ```

   **优点**：Emby 扫描音乐虚拟文件入库之后，能正常识别出音乐标签和时长

   **缺点**：调用 ffmpeg 解析远程参数，有一定的风控风险；扫描速度是三种方式中最慢的；Emby 源端口 8096 无法播放

### 其他配置

| 属性名                     | 描述                                                         | 示例值        |
| -------------------------- | :----------------------------------------------------------- | ------------- |
| `openlist`                 | ---                                                          | ---           |
| > `local-tree-gen`         | ---                                                          | ---           |
| >> `auto-remove-max-count` | 此配置相当于为本地目录树加了个保险措施，防止 openlist 存储挂载出现异常后，程序误以为远程文件被删除，而将本地已扫描完成的目录树清空的情况。<br /><br />具体配置值需以自己 openlist 的总文件数为参考（可留意首次全量扫描目录树后的日志输出），建议配置为总文件数的 3/4 左右大小，当程序即将要删除的文件数目超过这个数值时，会停止删除操作，并在日志中输出警告 | `6000`        |
| >> `refresh-interval`      | 本地目录树刷新间隔，单位：分钟                               | `60`          |
| >> `scan-prefixes`         | 指定 openlist 要扫描到本地的前缀列表，没有配置则默认全量扫描 | ---           |
| >> `ignore-containers`     | 忽略指定容器，避免触发源文件下载                             | `jpg,png,nfo` |

### 额外说明

1. 为了保持 10MB 大小的精简 Docker 镜像，ffmpeg 默认不会被添加到镜像中。首次将 `openlist.local-tree-gen.ffmpeg-enable` 配置设置为 `true` 并运行容器后，程序会自动初始化 ffmpeg 环境，请耐心等待下载完成，中途不要停止容器
2. 当检测出远程文件容器不在上述所说的三种生成方式任何一种之中时，程序的默认行为是将源文件下载到本地。比如在上述的例子中没有配置 `nfo` 文件格式，则远程的 `nfo` 文件会原封不动保存下载到本地。所以在使用本功能前请确认好所有的媒体大文件格式全都已经配置到了上述三种生成方式中

## 使用说明 对象存储(OSS) 重定向

> **⚠️ 魔改功能说明**：此功能为社区魔改版本，用于将 Emby 媒体请求直接重定向到对象存储（S3/OSS/COS等），适用于已将媒体文件存储在公共对象存储的场景。

### 功能特点

1. **直接重定向**：Emby 请求直接 302 重定向到对象存储 URL，无需经过 OpenList
2. **CDN Type-A 鉴权**：支持标准的 CDN Type-A 鉴权算法，自动生成带时效的 `auth_key` 参数
3. **源站验证**：重定向响应头携带自定义 API Key，用于对象存储源站的访问验证
4. **路径映射**：灵活的 Emby 本地路径到 OSS 路径的映射配置
5. **中文支持**：自动处理中文文件名的 URL 编码
6. **模式切换**：可通过配置在 OpenList 模式和 OSS 模式之间切换

### Type-A 鉴权算法说明

CDN Type-A 是一种常见的 URL 鉴权方式，算法如下：

```
鉴权 URL 格式：
http://DomainName/Path/file.mp4?auth_key={timestamp}-{rand}-{uid}-{md5hash}

参数说明：
- timestamp: Unix时间戳（10位），表示链接过期时间
- rand: 随机字符串（UUID格式，32位hex，无中划线），增强安全性
- uid: 用户标识，通常设置为 0
- md5hash: MD5签名值

MD5 计算方法：
sstring = "/Path/file.mp4-{timestamp}-{rand}-{uid}-{PrivateKey}"
md5hash = md5(sstring)

示例：
URI: /media/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv
timestamp: 1734567890
rand: 477b3bbc253f467b8def6711128c7bec
uid: 0
PrivateKey: your-cdn-private-key

sstring = "/media/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv-1734567890-477b3bbc253f467b8def6711128c7bec-0-your-cdn-private-key"
md5hash = md5(sstring)
auth_key = "1734567890-477b3bbc253f467b8def6711128c7bec-0-{md5hash}"

最终 URL:
https://s3.example.com/media/星际穿越%20(2014)/星际穿越%20(2014)%20-%202160p.mkv?auth_key=1734567890-477b3bbc253f467b8def6711128c7bec-0-{md5hash}
```

### 配置方式

在 `config.yml` 中添加 OSS 配置段：

```yaml
# 对象存储 OSS 配置
oss:
  enable: true                                    # 是否启用 OSS 重定向功能
  endpoint: https://s3.startspoint.com            # 对象存储访问域名（不包含 bucket）
  bucket: ""                                      # 存储桶名称（留空表示URL中不包含bucket）

  # Emby路径到OSS路径的映射关系
  # 格式: Emby本地路径前缀:OSS路径前缀
  path-mapping:
    - /movie:/media                               # 将 Emby 的 /movie 映射到 OSS 的 /media
    - /series:/media                              # 将 Emby 的 /series 也映射到 OSS 的 /media
    - /music:/music                               # 音乐路径映射

  # CDN Type-A 鉴权配置
  cdn-auth:
    enable: true                                  # 是否启用 CDN Type-A 鉴权
    private-key: your-cdn-private-key-here        # CDN 鉴权密钥（必填）
    ttl: 3600                                     # 鉴权URL有效期（秒），默认3600秒（1小时）
    uid: 0                                        # 用户ID，通常设置为 0
    use-random: true                              # 是否使用随机数增强安全性

  # 源站验证 API Key 配置
  api-key:
    enable: true                                  # 是否启用 API Key 验证
    header-name: X-Api-Key                        # 响应头名称
    key: your-api-key-here                        # API Key 值（必填）
```

### 配置说明

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `oss.enable` | bool | 是 | 启用后将使用 OSS 模式，禁用则使用 OpenList 模式 |
| `oss.endpoint` | string | 是 | 对象存储访问域名，如 `https://s3.example.com` |
| `oss.bucket` | string | 否 | 存储桶名称，如果 URL 不需要包含 bucket 则留空 |
| `oss.path-mapping` | []string | 是 | 路径映射列表，格式为 `Emby路径:OSS路径` |
| `oss.cdn-auth.enable` | bool | 否 | 是否启用 CDN 鉴权，不启用则生成无鉴权的直链 |
| `oss.cdn-auth.private-key` | string | 条件 | CDN 鉴权密钥，启用鉴权时必填 |
| `oss.cdn-auth.ttl` | int64 | 否 | 链接有效期（秒），默认 3600 |
| `oss.cdn-auth.uid` | string | 否 | 用户 ID，默认 "0" |
| `oss.cdn-auth.use-random` | bool | 否 | 是否使用随机数，默认 true |
| `oss.api-key.enable` | bool | 否 | 是否在响应头添加 API Key |
| `oss.api-key.header-name` | string | 否 | API Key 响应头名称，默认 `X-Api-Key` |
| `oss.api-key.key` | string | 条件 | API Key 值，启用时必填 |

### 使用示例

**场景**：媒体文件存储在 S3 兼容的对象存储，需要通过 CDN 鉴权访问

1. **配置对象存储信息**

```yaml
oss:
  enable: true
  endpoint: https://s3.startspoint.com
  bucket: ""
```

2. **配置路径映射**

假设你的 Emby 媒体库挂载路径为：
- `/movie` → 对应 OSS 的 `/media` 路径
- `/series` → 对应 OSS 的 `/media` 路径

```yaml
oss:
  path-mapping:
    - /movie:/media
    - /series:/media
```

3. **配置 CDN 鉴权**

```yaml
oss:
  cdn-auth:
    enable: true
    private-key: "your-actual-cdn-key-12345"
    ttl: 7200                    # 2小时有效期
    uid: "0"
    use-random: true
```

4. **配置 API Key（可选）**

```yaml
oss:
  api-key:
    enable: true
    header-name: X-Api-Key
    key: "your-api-key-67890"
```

### 工作流程

1. **客户端请求**：Emby 客户端请求视频流
   ```
   GET /emby/Videos/{id}/stream?api_key=xxx
   ```

2. **解析路径**：程序解析 Emby 媒体项的本地路径
   ```
   Emby 路径: /movie/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv
   ```

3. **路径映射**：根据配置映射到 OSS 路径
   ```
   OSS 路径: /media/星际穿越 (2014)/星际穿越 (2014) - 2160p.mkv
   ```

4. **生成鉴权 URL**：计算 Type-A 签名并构建完整 URL
   ```
   https://s3.startspoint.com/media/星际穿越%20(2014)/星际穿越%20(2014)%20-%202160p.mkv?auth_key=1734567890-477b3bbc253f467b8def6711128c7bec-0-abc123...
   ```

5. **302 重定向**：返回重定向响应
   ```
   HTTP/1.1 302 Temporary Redirect
   Location: https://s3.startspoint.com/media/...?auth_key=...
   X-Api-Key: your-api-key-67890
   ```

6. **客户端直连**：客户端直接从对象存储下载视频

### 注意事项

1. **启用 OSS 模式后，将不再使用 OpenList**，两种模式互斥
2. **不支持转码功能**：OSS 模式仅支持原画直链，不支持阿里云盘转码
3. **路径必须匹配**：确保 Emby 中配置的挂载路径与 `path-mapping` 中的前缀匹配
4. **中文路径**：程序会自动处理中文文件名的 URL 编码
5. **缓存时间**：生成的 OSS URL 缓存 10 分钟，减少重复计算
6. **错误处理**：如果路径映射失败，会根据 `emby.proxy-error-strategy` 配置决定回源或拒绝
7. **有效期设置**：`cdn-auth.ttl` 应大于视频播放时长，建议设置 3600 秒以上
8. **密钥安全**：`private-key` 和 `api-key.key` 请妥善保管，不要提交到公开仓库

### 故障排除

**问题1：路径映射失败**
- 检查日志：`无法映射 Emby 路径到 OSS 路径: xxx`
- 解决方法：确认 `path-mapping` 中包含该路径的前缀映射

**问题2：CDN 鉴权失败**
- 检查日志：查看生成的 `auth_key` 和 URL
- 解决方法：验证 `private-key` 是否正确，检查 CDN 服务商的鉴权算法文档

**问题3：客户端播放失败**
- 检查响应：使用 curl 测试重定向 URL 是否可访问
- 解决方法：确认对象存储的访问权限、网络连通性、鉴权参数

**问题4：中文路径无法访问**
- 检查 URL：确认中文字符是否正确编码为 `%E9%98%BF%E9%87%8C`
- 解决方法：程序会自动编码，如有问题请提 issue

### 与 OpenList 模式对比

| 特性 | OpenList 模式 | OSS 模式 |
|------|--------------|---------|
| 适用场景 | 网盘文件 | 对象存储文件 |
| 转码支持 | ✅ 支持阿里云盘转码 | ❌ 不支持 |
| 鉴权方式 | OpenList Token | CDN Type-A 鉴权 |
| 中转服务 | 需要 OpenList | 不需要 |
| 配置复杂度 | 中等 | 较低 |
| 性能 | 取决于 OpenList | 直连对象存储 |
| 缓存时间 | 10 分钟 | 10 分钟 |

## 请我喝杯 9.9💰 的 Luckin Coffee☕️

<img height="500px" src="assets/2024-11-05-09-57-45.jpg"></img>

## Star History

<a href="https://star-history.com/#AmbitiousJun/go-emby2openlist&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=AmbitiousJun/go-emby2openlist&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=AmbitiousJun/go-emby2openlist&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=AmbitiousJun/go-emby2openlist&type=Date" />
 </picture>
</a>
