FROM scratch

COPY --chmod=0700 /domain_exporter /domain_exporter

ENV COMMIT_SHORT_SHA=$CI_COMMIT_SHORT_SHA

ENTRYPOINT ["/domain_exporter"]