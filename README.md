# 基于开源MaxMindDB构建自定义IP数据库 2.0

使用Go重构socdb项目，相对于perl构建速度提高20倍以上，维护更加方便

# 使用

## 1. 下载项目
```bash
cd $GOPATH/src/
mkdir -p github.com/yanmengfei
git clone gitlab@git.socmap.org:yanmengfei/socip.git github.com/yanmengfei/
cd github.com/yanmengfei/socip
```

## 2. 修改配置文件
> cp config.yaml.sample config.yaml
```yaml
# 通用配置
password:  # 下载的zip解压密码
folder:   # 下载数据存放目录
ip_version: 4 # 构建的离线库版本
record_size: 24 # 数据库块大小[24, 28, 32]

# 基础数据库
basic:
  id:  # 某站下载数据源文件的 downloadId
  input: 'basic.txt'  # 输入文件名(对应下载并解压后的重命名)
  output: 'soc-geoip-v2.0.3.socdb' # 输出文件名
  database_type: 'SOC-GeoIP' # 数据库类型
  description:  # 描述信息
    - en: 'SOC ipv4 geographic information offline database'
    - cn: '赛欧思IP地理信息离线数据库'

# 场景数据库
scene:
  id:  # 某站下载数据源文件的 downloadId
  input: 'scene.txt' # 输入文件名(对应下载并解压后的重命名)
  output: 'soc-scene-v2.0.3.socdb' # 输出文件名
  database_type: 'SOC-Scene' # 数据库类型
  description:  # 描述信息
    - en: 'SOC ipv4 usage scenario offline database'
    - cn: '赛欧思IP使用场景离线数据库'
```

## 3. 构建数据库
```bash
# 编译项目
go build -o buildSocdb main.go
# 自动下载对应数据库并按照配置文件构建
# all: 同时构建basic和scene数据库
./buildSocdb -t [ all | basic | scene ]
```

## 4. 验证
```bash
go build -o validSocdb valid.go
# -n 验证的行数
./validSocdb -t [ all | basic | scene ] -n [10000 - 200000]
```

# 开发

```
./core/socdb  --> socdb构建核心代码，主要实现二叉搜索树的构建
./core/download.go --> 实现下载源数据文件
./core/parse.go --> 对源文件的解析
./core/utils --> IP转换相关工具
./global/config --> 加载配置文件和日志管理
./main.go --> socdb构建主逻辑
./valid.go --> socdb验证主逻辑
```