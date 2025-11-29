FROM public.ecr.aws/sam/build-go1.x:1.113

ARG UID=1000
ARG GID=1000

# 使用状況データをAWSに送信しない
ENV SAM_CLI_TELEMETRY=0

# 非特権ユーザの設定
RUN (getent passwd ${UID} && /usr/sbin/userdel -r $(getent passwd ${UID} | cut -d: -f1) || true) && \
  (getent group ${GID} || /usr/sbin/groupadd -g ${GID} nonroot) && \
  /usr/sbin/useradd -u ${UID} -g ${GID} -m -s /bin/bash nonroot

RUN go install github.com/air-verse/air@v1.63

# 非特権ユーザに切り替え
USER nonroot
WORKDIR /app
