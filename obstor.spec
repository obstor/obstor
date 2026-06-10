%define         tag     RELEASE.2020-11-25T22-36-25Z
%define         subver  %(echo %{tag} | sed -e 's/[^0-9]//g')
# git fetch https://github.com/obstor/obstor.git refs/tags/RELEASE.2020-11-25T22-36-25Z
# git rev-list -n 1 FETCH_HEAD
%define         commitid        91130e884b5df59d66a45a0aad4f48db88f5ca63
Summary:        High Performance, Kubernetes Native Object Storage.
Name:           obstor
Version:        0.0.%{subver}
Release:        1
Vendor:         PGG, Inc.
License:        Apache v2.0
Group:          Applications/File
Source0:        https://dl.pgg.net/packages/obstor/release/linux-amd64/archive/obstor.%{tag}
Source1:        https://raw.githubusercontent.com/obstor/obstor-service/master/linux-systemd/distributed/minio.service
URL:            https://pgg.net/
Requires(pre):  /usr/sbin/useradd, /usr/bin/getent
Requires(postun): /usr/sbin/userdel
BuildRoot:      %{tmpdir}/%{name}-%{version}-root-%(id -u -n)

## Disable debug packages.
%define         debug_package %{nil}

%description
Obstor is a High Performance Object Storage released under Apache License v2.0.
It is API compatible with Amazon S3 cloud storage service. Use Obstor to build
high performance infrastructure for machine learning, analytics and application
data workloads.

%pre
/usr/bin/getent group obstor-user || /usr/sbin/groupadd -r obstor-user
/usr/bin/getent passwd obstor-user || /usr/sbin/useradd -r -d /etc/obstor -s /sbin/nologin obstor-user

%install
rm -rf $RPM_BUILD_ROOT
install -d $RPM_BUILD_ROOT/etc/obstor/certs
install -d $RPM_BUILD_ROOT/etc/systemd/system
install -d $RPM_BUILD_ROOT/etc/default
install -d $RPM_BUILD_ROOT/usr/local/bin

cat <<EOF >> $RPM_BUILD_ROOT/etc/default/obstor
# Remote volumes to be used for Obstor server.
# Uncomment line before starting the server.
# OBSTOR_VOLUMES=http://node{1...6}/export{1...32}

# Root credentials for the server.
# Uncomment both lines before starting the server.
# OBSTOR_ROOT_USER=Server-Root-User
# OBSTOR_ROOT_PASSWORD=Server-Root-Password

OBSTOR_OPTS="--certs-dir /etc/obstor/certs"
EOF

install %{_sourcedir}/obstor.service $RPM_BUILD_ROOT/etc/systemd/system/obstor.service
install -p %{_sourcedir}/%{name}.%{tag} $RPM_BUILD_ROOT/usr/local/bin/obstor

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(644,root,root,755)
%attr(644,root,root) /etc/default/obstor
%attr(644,root,root) /etc/systemd/system/obstor.service
%attr(644,obstor-user,obstor-user) /etc/obstor
%attr(755,obstor-user,obstor-user) /usr/local/bin/obstor
