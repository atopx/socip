# encoding: utf-8

require "logstash/filters/base"
require "logstash/namespace"


class LogStash::Filters::SocGeoIP < LogStash::Filters::Base
  attr_accessor :socdb
  attr_accessor :scenedb
  config_name "socgeoip"
  config :geo_database, :validate => :path, :required => true
  config :scene_database, :validate => :path, :required => true
  config :source, :validate => :string, :required => true
  config :fields, :validate => :array
  config :target, :validate => :string, :default => 'socgeoip'

  public
  def register
    require 'maxmind/db'
    if @geo_database.nil? || !File.exists?(@geo_database)
      raise "You must specify 'database => ...' in your socgeoip filter (I looked for '#{@geo_database}')"
    end
    @logger.info("Using geoip database", :path => @geo_database)
    if @scene_database.nil? || !File.exists?(@scene_database)
      raise "You must specify 'database => ...' in your socgeoip filter (I looked for '#{@scene_database}')"
    end
    @logger.info("Using scene database", :path => @scene_database)

    begin
        @socdb = MaxMind::DB.new(@geo_database, mode: MaxMind::DB::MODE_MEMORY)
        @scenedb = MaxMind::DB.new(@scene_database, mode: MaxMind::DB::MODE_MEMORY)
    rescue Exception => e
        raise "Could'nt load the database file. (Do I have read access for '#{@geo_database}' or '#{@scene_database}'?)"
    end
  end # def register

  public
  def filter(event)
    return unless event.get(@source)
    begin
        eTarget = resolveIP(event.get(@source))
    rescue Exception => e
        @logger.error("IP Field contained invalid IP address or hostname", :e => e, :field => @source, :event => event)
    end
    event.set(@target, eTarget)
    filter_matched(event)
  end # def filter

  private
  def resolveIP(ip)
    data = @socdb.get(ip)
    scene_data = @scenedb.get(ip)
    if scene_data.nil?
      scene_data = Hash["scene"=>""]
    end
    if data.nil?
      @logger.info("#{ip} was not found in the database", :e => e, :field => ip, :event => event)
      data = Hash["accuracy"=>"", "city"=>"", "continent"=>"", "country"=>"", "district"=>"", "lat"=>"", "lng"=>"", "owner"=>"", "province"=>"", "radius"=>"", "scene"=>"", "source"=>"soc-geoip", "zipcode"=>""]
    end
    data["scene"] = scene_data["scene"]
    if !@fields.nil?
      data.each_key do |key|
        if !@fields.include? key
          data.delete(key)
        end
      end
    end
    return data
  end   # func resolveIP
end # class SocGeoIP

