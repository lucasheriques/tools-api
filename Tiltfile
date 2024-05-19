docker_build('tools-lucasfaria-dev', '.',
    dockerfile='./deployments/Dockerfile')
k8s_yaml(['./deployments/tools-lucasfaria.yaml', './deployments/gotenberg.yaml', './services/gotenberg.yaml'])
k8s_resource('tools-lucasfaria-dev', port_forwards=8080)
