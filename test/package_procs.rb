#!/usr/bin/env ruby

require 'json'

hub_username = ENV['DOCKERHUB_USERNAME']
hub_password = ENV['DOCKERHUB_PASSWORD']
jobs_path = ENV['PROCTOR_JOBS_PATH']
team_name = ENV['PROCTOR_JOB_TEAM_NAME'] || "test"
container_registry = ENV["PROCTOR_CONTAINER_REGISTRY"] || "docker.io/proctorscripts"
metadata_file_name = ENV['PROCTOR_METADATA_FILE_NAME'] || "metadata.json"

def run_cmd(cmd)
  puts cmd
  result = system(cmd)
  if !result
    puts "#{cmd} exited with non-zero code"
    exit 1
  end
end

def login(username, password)
  run_cmd("docker login -u #{username} -p #{password}")
end

for dir in Dir["#{jobs_path}/*/"]
  metadata_file = dir + '/' + metadata_file_name

  login(hub_username, hub_password)

  if File.exist?(metadata_file)
    file = File.read(metadata_file)
    data_hash = JSON.parse(file)
    image_name = "#{container_registry}/#{team_name}-#{data_hash['name']}:latest"

    Dir.chdir(dir) {
      puts "===== build and push image ====="
      run_cmd("docker build -t #{image_name} .")
      run_cmd("docker push #{image_name}")
    }
  else
    puts "#{dir} doesn't have metadata_file"
  end

end

