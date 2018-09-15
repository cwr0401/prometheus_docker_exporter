FROM scratch
COPY prometheus_docker_exporter_linux /prometheus_docker_exporter
EXPOSE 8000
ENTRYPOINT ["/prometheus_docker_exporter"]