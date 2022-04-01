FROM python:3
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
RUN apt-get update && apt-get install -y dnsutils netcat
RUN useradd -ms /bin/bash www
USER www
WORKDIR /code
COPY requirements.txt /code/
RUN pip install --user -r requirements.txt && rm -fr /home/www/.cache
COPY . /code/
CMD ["./hack/runserver.sh"]

