docker_build('tools-api', '.',
    dockerfile='./deployments/Dockerfile')
k8s_yaml(['./deployments/tools-api.yaml', './deployments/gotenberg.yaml', './services/gotenberg.yaml'])
k8s_resource('tools-api', port_forwards=8080)
