FROM node:16 AS node-build
WORKDIR /build
ADD . /build
RUN make web-assets

FROM golang:1.16 as go-build
COPY --from=node-build /build /build
WORKDIR /build
RUN make build

# FROM python:3.6-slim AS python-build
# RUN pip install --upgrade pip
# RUN ln -s /usr/local/bin/python /usr/bin/python
# RUN /usr/bin/python -m venv /venv
# RUN /venv/bin/pip install ansible ara

FROM registry.suse.com/suse/sle15:latest as sles
RUN zypper in -y python3-base


FROM registry.suse.com/suse/sle15:latest as python-build
RUN zypper in -y python3-base python3-pip
RUN /usr/bin/python3 -m venv /venv
RUN /venv/bin/pip install ansible ara

FROM scratch as sles-distroless-python3
COPY --from=sles /usr/lib64/libpython3.6m.so.1.0 /usr/lib64/libpython3.6m.so.1.0
COPY --from=sles /lib64/libpthread.so.0 /lib64/libpthread.so.0 
COPY --from=sles /lib64/libc.so.6 /lib64/libc.so.6 
COPY --from=sles /lib64/libdl.so.2  /lib64/libdl.so.2
COPY --from=sles /lib64/libutil.so.1 /lib64/libutil.so.1 
COPY --from=sles /lib64/libm.so.6 /lib64/libm.so.6
COPY --from=sles /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

COPY --from=sles /usr/bin/python3 /usr/bin/python3
COPY --from=sles /usr/lib64/python3.6 /usr/lib64/python3.6

COPY --from=sles /usr/lib64/libbz2.so.1 /usr/lib64/libbz2.so.1
COPY --from=sles /usr/lib64/libcrypt.so.1 /usr/lib64/libcrypt.so.1
COPY --from=sles /usr/lib64/libexpat.so.1 /usr/lib64/libexpat.so.1
COPY --from=sles /usr/lib64/libffi.so.7 /usr/lib64/libffi.so.7
COPY --from=sles /usr/lib64/liblzma.so.5 /usr/lib64/liblzma.so.5
COPY --from=sles /lib64/libm.so.6 /lib64/libm.so.6
COPY --from=sles /usr/lib64/libssl.so.1.1 /usr/lib64/libssl.so.1.1
COPY --from=sles /lib64/libz.so.1 /lib64/libz.so.1
COPY --from=sles /lib64/libgcc_s.so.1 /lib64/libgcc_s.so.1

FROM sles-distroless-python3
# COPY --from=sles-distroless-python3 /usr /usr
# COPY --from=sles-distroless-python3 /lib64 /lib64

COPY --from=python-build /venv /venv
ENV PATH="/venv/bin:$PATH"
ENV PYTHONPATH=/venv/lib/python3.6/site-packages
COPY --from=go-build /build/trento /app/trento

EXPOSE 8080/tcp
ENTRYPOINT ["/app/trento"]
