FROM scratch
ADD bin /
CMD ["/buildwatcher"]

# https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/