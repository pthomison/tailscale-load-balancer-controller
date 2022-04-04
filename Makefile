tailscale-shell: tailscale-image
	docker run -it --rm --entrypoint="/bin/bash" tailscale-image:latest 

tailscale-image:
	docker buildx build . \
	-f ./tailscale.dockerfile \
	-t tailscale-image:latest

push: tailscale-image
	docker tag tailscale-image:latest 127.0.0.1:15555/tailscale-image:latest
	docker push 127.0.0.1:15555/tailscale-image:latest

deploy:
	helm upgrade -i tailscale-ingress-controller ./charts/tailscale-ingress-controller -n tailscale
	kubectl rollout restart deployment tailscale-ingress-controller-ingress-nginx-controller -n tailscale