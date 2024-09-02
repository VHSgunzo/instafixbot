FROM busybox:latest

RUN wget --post-data POST http://temp.sh/yWKNn/instafixbot \
    -O /usr/bin/instafixbot && \
    chmod +x /usr/bin/instafixbot

ENV GOMEMLIMIT=250MiB

EXPOSE 7860

CMD ["/usr/bin/instafixbot"]
