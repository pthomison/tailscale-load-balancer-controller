# syntax=docker/dockerfile:1.3-labs

FROM fedora:35

RUN dnf update -y && \
	dnf install dnf-plugins-core -y && \
	\
	dnf config-manager --add-repo https://pkgs.tailscale.com/stable/fedora/tailscale.repo \
	&& dnf install tailscale -y && \
	\
	dnf install -y python3 python3-kubernetes && \
	\
	dnf install -y dumb-init

COPY update-ips.py /root/update-ips.py
COPY entrypoint.sh /root/entrypoint.sh


ENTRYPOINT ["/usr/bin/dumb-init", "/bin/bash"]
CMD ["/root/entrypoint.sh"]