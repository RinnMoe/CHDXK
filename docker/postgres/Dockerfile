ARG PG_MAJOR=17

FROM postgres:${PG_MAJOR} AS builder

RUN sed -i 's|http://deb.debian.org/|http://mirrors.tuna.tsinghua.edu.cn/|g' /etc/apt/sources.list.d/debian.sources

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        git \
        ca-certificates \
        cmake \
        postgresql-server-dev-${PG_MAJOR} \
    && rm -rf /var/lib/apt/lists/*

RUN git clone --depth 1 https://github.com/jaiminpan/pg_jieba.git /pg_jieba \
    && cd /pg_jieba \
    && git submodule update --init --recursive \
    && mkdir build && cd build \
    && cmake -DPostgreSQL_TYPE_INCLUDE_DIR=/usr/include/postgresql/${PG_MAJOR}/server .. \
    && make && make install \
    && cd / && rm -rf /pg_jieba

FROM postgres:${PG_MAJOR}

ARG PG_MAJOR

COPY --from=builder /usr/lib/postgresql/${PG_MAJOR}/lib/pg_jieba.so \
    /usr/lib/postgresql/${PG_MAJOR}/lib/pg_jieba.so

COPY --from=builder /usr/share/postgresql/${PG_MAJOR}/extension/pg_jieba* \
    /usr/share/postgresql/${PG_MAJOR}/extension/

COPY --from=builder /usr/share/postgresql/${PG_MAJOR}/tsearch_data/ \
    /usr/share/postgresql/${PG_MAJOR}/tsearch_data/

COPY initdb /docker-entrypoint-initdb.d/