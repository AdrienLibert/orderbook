FROM python:3.12.1-bullseye

RUN mkdir /app
COPY ./kafka_init /app/
RUN mkdir /app/dist
COPY ./drgn/dist/ /app/dist

WORKDIR /app

RUN pip3 install uv~=0.5
RUN python3 -m uv pip install -r requirements.txt --find-links file://${PWD}/dist/
RUN touch /app/task.conf

CMD ["python3", "entrypoint.py"]