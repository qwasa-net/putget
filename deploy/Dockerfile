FROM scratch
LABEL name=putget

EXPOSE 18801

WORKDIR /

ADD putget /

VOLUME "/files"

ENV PUTGET_FILES_ROOT="/files/"
ENV PUTGET_DB_PATH="/files/putget.sqlite"
ENV PUTGET_BIND_ADDRESS="0.0.0.0:18801"
ENV PUTGET_URL_ROOT="/"

CMD ["/putget"]