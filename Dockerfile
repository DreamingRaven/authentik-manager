ARG PKG_DIR="/app"
ARG SRC_DIR="docs/server"
ARG DOC_DIR="docs"
ARG EXPOSE_PORT="8080"
ARG EXPOSE_PORT_PROTOCOL="tcp"
ARG USERNAME="archer"


FROM golang:1.21.5-alpine as gopiler
ARG SRC_DIR
ARG PKG_DIR
RUN mkdir -p ${PKG_DIR}
COPY ${SRC_DIR} ${PKG_DIR}
WORKDIR ${PKG_DIR}
RUN go version && \
    go mod tidy && \
    mkdir -p ${PKG_DIR}/build && \
    go build -o build/server.go


FROM python as docpiler
ARG PKG_DIR
ARG DOC_DIR
RUN mkdir -p /app/docs && \
    apt update && \
    apt install -y tree
COPY ${DOC_DIR}/requirements.txt ${PKG_DIR}/${DOC_DIR}/requirements.txt
WORKDIR ${PKG_DIR}/${DOC_DIR}
RUN pip3 install -r requirements.txt
COPY .git ${PKG_DIR}/${DOC_DIR}
COPY ${DOC_DIR} ${PKG_DIR}/${DOC_DIR}
RUN make html
RUN tree ${PKG_DIR}/${DOC_DIR}/build/


FROM golang:1.21.5-alpine as runner
ARG PKG_DIR
ARG DOC_DIR
ARG USERNAME
ARG EXPOSE_PORT
ARG EXPOSE_PORT_PROTOCOL
COPY --from=gopiler ${PKG_DIR}/build ${PKG_DIR}
COPY --from=docpiler ${PKG_DIR}/${DOC_DIR}/build/html ${PKG_DIR}/static
COPY --from=gopiler ${PKG_DIR}/static ${PKG_DIR}/static
RUN apk add shadow tree coreutils bash && \
    chmod 544 ${PKG_DIR}/server.go && \
    useradd ${USERNAME} && \
    chown -R ${USERNAME}:${USERNAME} ${PKG_DIR}
USER ${USERNAME}
WORKDIR ${PKG_DIR}
RUN tree .
EXPOSE ${EXPOSE_PORT}/${EXPOSE_PORT_PROTOCOL}
CMD ./server.go

