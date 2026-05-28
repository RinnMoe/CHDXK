docker build -f api.Dockerfile -t registry.cn-hongkong.aliyuncs.com/jcourse/jcourse_api:latest .
docker build -f taskworker.Dockerfile -t registry.cn-hongkong.aliyuncs.com/jcourse/jcourse_taskworker:latest .

docker push registry.cn-hongkong.aliyuncs.com/jcourse/jcourse_api:latest
docker push registry.cn-hongkong.aliyuncs.com/jcourse/jcourse_taskworker:latest