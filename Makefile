tailscale-shell: tailscale-image
	docker run -it --rm --entrypoint="/bin/bash" tailscale-image:latest 

tailscale-image:
	docker buildx build . \
	-f ./tailscale.dockerfile \
	-t tailscale-image:latest

push: tailscale-image
	docker tag tailscale-image:latest 127.0.0.1:15555/tailscale-image:latest
	docker push 127.0.0.1:15555/tailscale-image:latest

deploy: push
	kubectl apply -f ./namespace.yaml
	helm upgrade -i nginx-ingress ingress-nginx/ingress-nginx -n tailscale -f ./nginx-ingress.values.yaml
	kubectl apply -f ./rbac.yaml
	kubectl rollout restart deployment/nginx-ingress-ingress-nginx-controller -n tailscale

clean: cluster-delete registry-delete