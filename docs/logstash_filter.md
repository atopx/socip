# Logstash插件开发

> 本文基于Docker运行的ubuntu:18.04为例
>
> ```bash
> docker run -it ubuntu:18.04 bash
> sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list
> apt-get update
> apt-get install wget -y
> ```
>

## 1. 环境安装

### 安装JDK1.8

```bash
apt install openjdk-1.8-jre -y
```

### 安装Logstash

```bash
apt-get install gnupg2 -y
wget -qO - https://artifacts.elastic.co/GPG-KEY-elasticsearch | apt-key add -
apt-get install apt-transport-https -y
echo "deb https://artifacts.elastic.co/packages/7.x/apt stable main" | tee -a /etc/apt/sources.list.d/elastic-7.x.list
apt-get update
apt-get install logstash -y
```

### 安装jruby

```bash
cd /opt
wget https://s3.amazonaws.com/jruby.org/downloads/9.2.11.1/jruby-bin-9.2.11.1.tar.gz
tar -xvf jruby-bin-9.2.11.1.tar.gz
mv jruby-9.2.11.1 jruby
echo "PATH=$PATH:/opt/jruby/bin:" >> ~/.bashrc
source ~/.bashrc
```

> 检测是否安装成功： 
>
> ```bash
> $ jruby -v
> jruby 9.1.13.0 (2.3.3) 2017-09-06 8e1c115 OpenJDK 64-Bit Server VM 25.91-b14 on 1.8.0_91-b14 +jit [linux-x86_64]
> ```

### 安装Bundler

```bash
# 安装
jgem install bundler
# 将资源修改为国内
bundle config mirror.https://rubygems.org https://gems.ruby-china.org
```

## 2. 生成插件模版

```bash
/usr/share/logstash/bin/logstash-plugin generate --type filter --name socgeoip  --path /opt
```

> --type: 插件类型
>
> --name: 插件名称
>
> --path: 插件生成路径

```bash
 Creating /opt/logstash-filter-socgeoip
	 create logstash-filter-socgeoip/docs/index.asciidoc
	 create logstash-filter-socgeoip/lib/logstash/filters/socgeoip.rb
	 create logstash-filter-socgeoip/spec/filters/socgeoip_spec.rb
	 create logstash-filter-socgeoip/spec/spec_helper.rb
	 create logstash-filter-socgeoip/README.md
	 create logstash-filter-socgeoip/CONTRIBUTORS
	 create logstash-filter-socgeoip/DEVELOPER.md
	 create logstash-filter-socgeoip/logstash-filter-socgeoip.gemspec
	 create logstash-filter-socgeoip/CHANGELOG.md
	 create logstash-filter-socgeoip/Rakefile
	 create logstash-filter-socgeoip/LICENSE
	 create logstash-filter-socgeoip/Gemfile
```



### 3. 编写插件代码

```bash
# 插件目录结构
-- logstash-filter-socgeoip
    |-- CHANGELOG.md	# 更新日志
    |-- CONTRIBUTORS	# 贡献
    |-- DEVELOPER.md	# 开发者
    |-- Gemfile	# 环境相关（可以修改source 'https://gems.ruby-china.com/'）
    |-- LICENSE	# 开源协议
    |-- README.md
    |-- Rakefile	# 似于C的make, 用来定义和执行任务
    |-- docs	# 直接删了省事
    |   -- index.asciidoc
    |-- src
    |	-- main
    |		-- java
    | 		-- org.logstash.filters
    |				-- Fields.java
    |				-- SocIPFilter.java
    |-- lib	# 真正有用的东西
    |   -- logstash
    |       -- filters
    |           -- socgeoip.rb	# 编写插件逻辑代码
    |-- logstash-filter-socgeoip.gemspec	# 插件元数据
    -- spec	# 测试代码
        |-- filters
        |   -- socgeoip_spec.rb	# 编写插件测试代码
        -- spec_helper.rb
```

#### 逻辑代码：

> lib/logstash/filters/socgeoip.rb

```ruby
# encoding: utf-8

# 引包
require "logstash/filters/base"
require "logstash/namespace"
require 'jar_dependencies'

# 由于该项目调用Java代码，需要引入Java所使用的三方包
require_jar('com.maxmind.db', 'maxmind-db', '1.2.2')

# 导入自定义编写的Java代码（Java代码可以放在'src/main/java/org/logstash/filters'路径下）
require_jar('org.logstash.filters', 'logstash-filter-geoip', '6.0.0')

# Socgeoip类，继承 LogStash::Filters::Base 类
class LogStash::Filters::Socgeoip < LogStash::Filters::Base
    # 插件名
    config_name "socgeoip"

  	# 定义类成员变量(logstash.conf中传入过来的变量字段)
    config :database, :validate => :path, :required => true
    config :default_database_type, :validate => ["SOC-GeoIP","SOC-Scene"], :default => "SOC-GeoIP"
    config :source, :validate => :string, :required => true
    config :fields, :validate => :array
    config :target, :validate => :string, :default => 'socgeoip'
    config :tag_on_failure, :validate => :array, :default => ["_socgeoip_lookup_failure"]

  	# 初始化方法
    public
    def register
        if @database.nil?
            @database = ::Dir.glob(::File.join(::File.expand_path("../../../vendor/", ::File.dirname(__FILE__)),"GeoLite2-#{@default_database_type}.mmdb")).first

            if @database.nil? || !File.exists?(@database)
                raise "You must specify 'database => ...' in your geoip filter (I looked for '#{@database}')"
            end
        end
        @logger.info("Using geoip database", :path => @database)
       	# 创建实例GeoIPFilter（Java Class）
        @socipfilter = org.logstash.filters.SocIPFilter.new(@source, @target, @fields, @database)
    end

  	# 核心逻辑方法
    public
    def filter(event)
        return unless filter?(event)
      	# 此处调用的是Java方法，一般逻辑简单不需要Java
        if @socipfilter.handleEvent(event)
          	# 查询成功，传入下一个过滤器
            filter_matched(event)
        else
          	# 查询失败，调用tag_unsuccessful_lookup方法
            tag_unsuccessful_lookup(event)
        end
    end
		
    # 定义过滤失败的方法
    def tag_unsuccessful_lookup(event)
        @logger.debug? && @logger.debug("IP #{event.get(@source)} was not found in the database", :event => event)
        @tag_on_failure.each{|tag| event.tag(tag)}
    end
end
```



#### 测试代码

> spec/filters/socgeoip_spec.rb

```ruby
# encoding: utf-8

require "logstash/devutils/rspec/spec_helper"
require "logstash/filters/socgeoip"

describe LogStash::Filters::SocGeoIP do
  describe "default" do
    config <<-CONFIG
      filter {
        socgeoip {
          source => "ip"
	      target => "socgeoip"
          database => "/opt/soc-geoip-2020.7.1.socdb"
        }
      }
    CONFIG

    sample("ip" => "117.158.142.120") do
      insist { subject }.include?("socgeoip")
      insist { subject["socgeoip"]["city"] } == "郑州市"
      insist { subject["socgeoip"]["scene"] } == "企业专线"

      expected_fields = %w(accuracy continent country province city
                           district lat lng owner radius scene source zipcode)

      expected_fields.each do |f|
        insist { subject["socgeoip"] }.include?(f)
      end
    end
  end
end
```

#### 元数据修改

> logstash-filter-socgeoip.gemspec

```ruby
Gem::Specification.new do |s|

  s.name            = 'logstash-filter-socgeoip'
  s.version         = '1.0.0'
  s.licenses        = ['Apache-2.0']
  s.summary         = "Adds geographical information about an IP address."
  s.description     = "This gem is a logstash plugin required to be installed on top of the Logstash core pipeline using $LS_HOME/bin/plugin install gemname. This gem is not a stand-alone program"
  s.authors         = ["itmeng"]
  s.email           = 'yanmengfei@socmap.net'
  s.homepage        = "http://www.itmeng.top"
  s.require_paths = ["lib"]

  # Files
  s.files = Dir['lib/**/*','spec/**/*','vendor/**/*','*.gemspec','*.md','Gemfile','LICENSE','NOTICE.TXT']

  # Tests
  s.test_files = s.files.grep(%r{^(test|spec|features)/})
  s.requirements << "jar 'org.apache.logging.log4j:log4j-core', '2.6.2'"

  # Special flag to let us know this is actually a logstash plugin
  s.metadata = { "logstash_plugin" => "true", "logstash_group" => "filter" }

  # Gem dependencies
  s.add_runtime_dependency 'jar-dependencies'
  s.add_runtime_dependency "logstash-core-plugin-api", "~> 2.0"
  s.add_runtime_dependency 'maxmind-db', '~> 1.1.0'
  s.add_development_dependency 'logstash-devutils'
end
```



### 4. 调试 

```bash
# 安装bundler
jgem install bundle

# 安装依赖
bundle install

# 单元测试
bundle exec rspec

# build Gem
jgem build logstash-filter-socgeoip.gemspec

# 更换国内gems镜像源
sed -i 's/rubygems.org/gems.ruby-china.com/' /usr/share/logstash/Gemfile

# 手动安装(执行该命令时，可能会出现卡死现象，ctrl+c(一下) 终止后，安装正常)
logstash-plugin install logstash-filter-socgeoip-1.0.0.gem

# 此时可以在本机进行实例测试，确认无误后进行打包
```

### 5. 打包

```bash
/usr/share/logstash/bin/logstash-plugin prepare-offline-pack logstash-filter-socgeoip
```

>  Offline package created at: /opt/logstash-offline-plugins-7.6.2.zip
>  You can install it with this command `bin/logstash-plugin install file:///opt/logstash-offline-plugins-7.6.2.zip`


### 6. 安装

```bash
/usr/share/logstash/bin/logstash-plugin install https://git.socmap.org/soc/smis/logstash_filter_socgeoip/uploads/logstash-offline-plugins-7.6.2.zip
```

> Downloading file: https://git.socmap.org/soc/smis/logstash_filter_socgeoip/uploads/logstash-offline-plugins-7.6.2.zip
>
> Installing file: logstash-offline-plugins-7.6.2-.zip
> Install successful

### 8. 使用

> logstash.conf

```ruby
input {
    file {
        path => "/opt/logstash/ips.list"
        start_position => "beginning"
        discover_interval => 15
        stat_interval => 1
        sincedb_path => "/opt/logstash/logs/.sincedb"
        sincedb_write_interval => 10
        mode => "read"
        file_completed_action => "log"
        file_completed_log_path => "/opt/logstash/logs/input.log"
        codec => json { charset => "UTF-8" }
    }
}

filter {
    mutate {
        remove_field => "@timestamp"
        remove_field => "@version"
        remove_field => "host"
        remove_field => "path"
    }

    socgeoip {
        source => "ip"
        target => "socgeoip"
        fields => ["country", "owner", "province", "city", "scene"]
        geo_database => ["/opt/soc-geoip-2020.7.1.socdb"]
        scene_database => ["/opt/soc-scene-2020-3.1.socdb"]
    }

}

output {
    stdout {
        codec => rubydebug
    }
}
```

> Output：
>
> ```ruby
> {
>     "socgeoip" => {
>            "scene" => "",
>          "country" => "中国",
>             "city" => "",
>            "owner" => "中国联通",
>         "province" => "山东省"
>     },
>           "ip" => "124.130.230.229"
> }
> {
>     "socgeoip" => {
>            "scene" => "企业专线",
>          "country" => "中国",
>             "city" => "郑州市",
>            "owner" => "中国移动",
>         "province" => "河南省"
>     },
>           "ip" => "117.158.142.120"
> }
> ```
