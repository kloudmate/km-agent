# 1. Create a Dockerfile (e.g., Dockerfile.build)
FROM rockylinux:8
RUN dnf update -y && \
    dnf install -y golang make rpm-build git gcc glibc-devel # Add other build deps
WORKDIR /build
# Copy source code if needed, or mount it

# 2. Build the container image
docker build -t kmagent-builder -f Dockerfile .

# 3. Run the build inside the container (mounting your source)
docker run --rm -v $(pwd):/build kmagent-builder make package-linux-rpm