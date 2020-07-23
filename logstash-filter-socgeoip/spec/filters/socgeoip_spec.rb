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
          database => "/Users/meng/soc-geoip-2020.7.1.socdb"
        }
      }
    CONFIG

    sample("ip" => "117.158.142.120") do
      insist { subject }.include?("socgeoip")
      insist { subject["socgeoip"]["city"] } == "郑州市"

      expected_fields = %w(accuracy continent country province city
                           district lat lng owner radius scene source zipcode)

      expected_fields.each do |f|
        insist { subject["socgeoip"] }.include?(f)
      end
    end
  end
end
