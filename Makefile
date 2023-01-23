python-shell: tailscale-image
	docker run -it --rm --entrypoint="/bin/bash" tailscale-ingress-controller:latest

python-image:
	cd python && \
	docker buildx build . \
	-t tailscale-ingress-controller:latest

python-push: tailscale-image
	docker tag tailscale-ingress-controller:latest 127.0.0.1:15555/tailscale-ingress-controller:latest
	docker push 127.0.0.1:15555/tailscale-ingress-controller:latest

deploy:
	helm upgrade -i tic-debug ./charts/tailscale-ingress-controller -n tailscale -f ./dev.yaml
	kubectl rollout restart deployment tic-debug-nic-controller -n tailscale