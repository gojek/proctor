#!/usr/bin/env ruby

require 'json'
require 'net/http'
require 'uri'

proctor_uri = ENV['PROCTOR_URI']

jobs_path = ENV['PROCTOR_JOBS_PATH']
team_name = ENV['PROCTOR_JOB_TEAM_NAME'] || "test"
container_registry = ENV["PROCTOR_CONTAINER_REGISTRY"] || "docker.io/proctorscripts"
metadata_file_name = ENV['PROCTOR_METADATA_FILE_NAME'] || "metadata.json"

sleep(2)

jobs = []

for dir in Dir["#{jobs_path}/*/"]
  metadata_file = dir +  metadata_file_name
  puts "Processing #{metadata_file}"

  if File.exist?(metadata_file)
    file = File.read(metadata_file)
    data_hash = JSON.parse(file)
    image_name = "#{container_registry}/#{team_name}-#{data_hash['name']}:latest"
    data_hash['image_name'] = image_name
    jobs << data_hash
  end
end

uri = URI.parse(proctor_uri)
header = {"Content-Type" => "application/json"}

# Create the HTTP objects
http = Net::HTTP.new(uri.host, uri.port)
request = Net::HTTP::Post.new(uri.request_uri, header)
request.body = jobs.to_json

# Send the request
puts "making req with body #{request.body}"
response = http.request(request)

if response.code == "201"
  puts 'Updated proctor metadata'
else
  puts 'Something went wrong while updating proctor metadata! Response from proctor:'
  puts response
  exit 1
end

