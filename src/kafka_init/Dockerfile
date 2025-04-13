FROM python:3.12.1-bullseye

RUN mkdir /app
COPY . /app/

WORKDIR /app

RUN pip3 install uv~=0.5
RUN python3 -m uv pip install -r requirements.txt
RUN touch /app/task.conf

CMD ["python3", "entrypoint.py"]