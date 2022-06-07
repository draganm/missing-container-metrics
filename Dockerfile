FROM scratch
ARG TARGETARCH
COPY missing-container-metrics-${TARGETARCH}-linux/missing-container-metrics /missing-container-metrics
EXPOSE 3001
ENTRYPOINT ["/missing-container-metrics"]
