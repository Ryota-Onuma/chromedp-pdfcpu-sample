DOCKER_IMAGE_NAME = "chromedp-pdfcpu-sample"
.PHONY bash:
bash:
	docker build --tag $(DOCKER_IMAGE_NAME) -f Dockerfile .
	docker run -it --rm -p 3000:3000 -v .:/home/app/workspace -w /home/app/workspace $(DOCKER_IMAGE_NAME) bash
