FROM public.ecr.aws/sam/build-go1.x:1.113

ARG UID=1000
ARG GID=1000

# 使用状況データをAWSに送信しない
ENV SAM_CLI_TELEMETRY=0

# 非特権ユーザの設定
RUN (getent passwd ${UID} && /usr/sbin/userdel -r $(getent passwd ${UID} | cut -d: -f1) || true) && \
  (getent group ${GID} || /usr/sbin/groupadd -g ${GID} nonroot) && \
  /usr/sbin/useradd -u ${UID} -g ${GID} -m -s /bin/bash nonroot

ARG DOCKER_GID=2375

# 非特権ユーザにDockerへのアクセス権限を与える
RUN (getent group ${DOCKER_GID} || /usr/sbin/groupadd -g ${DOCKER_GID} docker) && \
  /usr/sbin/usermod -aG $(getent group ${DOCKER_GID} | cut -d: -f1) nonroot

# dockerクライアントのインストール
COPY --from=docker:28.2-cli /usr/local/bin/docker /usr/local/bin/docker

# entrypointの上書き
COPY ./server-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# 非特権ユーザに切り替え
USER nonroot
WORKDIR /app

CMD ["sam", "local", "start-lambda", "--host", "0.0.0.0", "--container-host-interface", "0.0.0.0", "--config-env", "local"]
