Name:           nibble
Version:        0.1.0
Release:        1%{?dist}
Summary:        Fast local network scanner with terminal UI

License:        MIT
URL:            https://github.com/backendsystems/nibble
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.21

%description
Nibble is a CLI tool for local network scanning that focuses on speed
and ease of use. Select a network interface, and Nibble scans your local
subnet, listing hosts, hardware manufacturer, open ports and services.

Features include lightning fast scans, hardware vendor lookup, port
banner reading, scan history, and a terminal UI built with Bubble Tea.

%prep
%autosetup

%build
go build -mod=vendor -ldflags="-X main.version=%{version}" -o nibble .

%install
install -D -m 0755 nibble %{buildroot}%{_bindir}/nibble

%files
%license LICENSE
%doc README.md
%{_bindir}/nibble

%changelog
* Wed Mar 18 2026 saberd <mail@saberd.com> - 0.1.0-1
- Initial release
