# syntax=docker/dockerfile:1.3-labs

FROM pthomison/tailscale:v1.24.2

COPY requirements.txt /root/requirements.txt

RUN pip3 install -r /root/requirements.txt

COPY update-ips.py /root/update-ips.py
COPY entrypoint.sh /root/entrypoint.sh


ENTRYPOINT ["/usr/bin/dumb-init", "/bin/zsh"]
CMD ["/root/entrypoint.sh"]