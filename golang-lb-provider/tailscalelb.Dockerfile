# syntax=docker/dockerfile:1.3-labs

FROM tailscale/tailscale:stable

RUN apk add curl bash

RUN <<EOF
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
EOF

RUN echo <<EOF > /ip-monitor-entrypoint.sh
#!/usr/bin/env bash

while true
do
    IP="$(tailscale --socket /tmp/tailscaled.sock ip --4)"
    echo $IP
    sleep 1
done

EOF

RUN chmod +x /ip-monitor-entrypoint.sh

# ENTRYPOINT [ "/entrypoint.sh" ]