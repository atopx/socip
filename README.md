
# 基于MaxMind的mmdb构建自定义socip库

> MaxMindDB 原理详细文档：[MaxMind原理详解](./docs/maxmind_readme.md)
> 
> 基于geoip的Logstash插件开发文档：[Logstash-filter-socgeoip开发文档](./docs/logstash_filter.md)

> 警告：已删除所有数据，只提供技术分享，使用者不得用于商业！

  - perl_build => Perl 构建socdb
  - python_build => Python 构建socdb 
  - logstash_filter_socgeoip => 基于socdb的logstash-filter使用插件


## 基于Perl构建socdb(生产使用)
  > 需要perl5环境

1. 安装依赖库(ubuntu 18.04)
    
  > 替换`current_user`
  ```bash
  sudo apt-get install gcc g++ make
  sudo cpan install MaxMind::DB::Common
  sudo cpan install Data::IEEE754
  sudo cpan install Net::Works::Network
  sudo cpan install Net::CIDR
  sudo cpan install Net::CIDR::Lite
  sudo chown -R current_user:current_user ~/.cpan/ 
  ```


2. 构建db
  ```bash
  perl perl_build/socip_build.pl source.csv output.socdb
  ```


## 基于Python构建socdb(构建算法和搜索算法研究)
  > 需要python3.5+环境

- 编译成socdb
  ```bash
  cd python_build && python3 main.py input.csv output.socdb
  ```

