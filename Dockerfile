FROM scratch
COPY prometheus_docker_exporter_linux /prometheus_docker_exporter
EXPOSE 8000
CMD ["/prometheus_docker_exporter"]