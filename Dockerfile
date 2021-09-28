FROM adoptopenjdk/openjdk11:jdk-11.0.11_9-debian

RUN apt-get clean
RUN rm -rf /var/lib/apt/lists/*
RUN apt-get update && apt-get -y install git-lfs vim wget curl git gcc

COPY go /usr/local/
RUN echo "export GOROOT=/home/jenkins/go" >> /etc/profile
RUN echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
RUN echo “dash dash/sh boolean false” | debconf-set-selections
RUN DEBIAN_FRONTEND=noninteractive dpkg-reconfigure dash
RUN source /etc/profile
COPY ./modify-once-hpa/* /home/
#RUN go build -ldflags "-s -w" -a -installsuffix cgo -o app .
WORKDIR /home/
RUN go build -o app main.go
RUN chmod 777 app

ARG VERSION=4.9
ARG user=jenkins
ARG group=jenkins
ARG uid=1000
ARG gid=1000
RUN groupadd -g ${gid} ${group}
RUN useradd -c "Jenkins user" -d /home/${user} -u ${uid} -g ${gid} -m ${user}
LABEL Description="This is a base image, which provides the Jenkins agent executable (agent.jar)" Vendor="Jenkins project" Version="${VERSION}"
ARG AGENT_WORKDIR=/home/${user}/agent
#RUN apt-get update && apt-get -y install git-lfs vim wget curl git
RUN curl --create-dirs -fsSLo /usr/share/jenkins/agent.jar https://repo.jenkins-ci.org/public/org/jenkins-ci/main/remoting/${VERSION}/remoting-${VERSION}.jar \
  && chmod 755 /usr/share/jenkins \
  && chmod 644 /usr/share/jenkins/agent.jar \
  && ln -sf /usr/share/jenkins/agent.jar /usr/share/jenkins/slave.jar
USER ${user}
ENV AGENT_WORKDIR=${AGENT_WORKDIR}
RUN mkdir /home/${user}/.jenkins && mkdir -p ${AGENT_WORKDIR}
VOLUME /home/${user}/.jenkins
VOLUME ${AGENT_WORKDIR}
WORKDIR /home/${user}
COPY jenkins-slave /usr/local/bin/jenkins-slave
ENTRYPOINT ["jenkins-slave"]
