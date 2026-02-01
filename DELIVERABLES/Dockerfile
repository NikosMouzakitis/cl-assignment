FROM quay.io/centos/centos:stream8

# -----------------------------
# repos fix: Stream 8 is archived
# Disabling default .repo files and 
# using vault.centos.org.
# -----------------------------
RUN set -eux; \
    mkdir -p /etc/yum.repos.d/backup; \
    mv -f /etc/yum.repos.d/*.repo /etc/yum.repos.d/backup/ || true; \
    cat > /etc/yum.repos.d/centos-stream8-vault.repo <<'EOF'
[baseos]
name=CentOS Stream 8 - BaseOS (vault)
baseurl=https://vault.centos.org/centos/8-stream/BaseOS/$basearch/os/
enabled=1
gpgcheck=0

[appstream]
name=CentOS Stream 8 - AppStream (vault)
baseurl=https://vault.centos.org/centos/8-stream/AppStream/$basearch/os/
enabled=1
gpgcheck=0

[extras]
name=CentOS Stream 8 - Extras (vault)
baseurl=https://vault.centos.org/centos/8-stream/extras/$basearch/os/
enabled=1
gpgcheck=0

[powertools]
name=CentOS Stream 8 - PowerTools (vault)
baseurl=https://vault.centos.org/centos/8-stream/PowerTools/$basearch/os/
enabled=1
gpgcheck=0
EOF

# ---------------------------------
# Installing tools for the build.
# ---------------------------------
RUN set -eux; \
    dnf -y update; \
    dnf -y install \
        dnf-plugins-core \
        'dnf-command(builddep)' \
        rpm-build rpmdevtools redhat-rpm-config \
        git curl wget \
        make gcc gcc-c++ \
        bc bison flex \
        elfutils-libelf-devel openssl-devel \
        perl python3 \
        tar xz gzip bzip2 \
        rsync diffutils \
        which file patch findutils; \
    dnf clean all

# -----------------------------------
# Create a non-root user
# -----------------------------------
RUN useradd -m -u 1000 builder && \
    su - builder -c "rpmdev-setuptree"

USER builder
WORKDIR /home/builder

# Initial cmd
CMD ["/bin/bash"]
