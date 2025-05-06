%define appname kmagent
%define scriptdir %{_libexecdir}/%{appname}
%define confdir /etc/%{appname}
%global _unitdir /usr/lib/systemd/system

Name: %{appname}
Version: 1.0.0
Release: 1%{?dist}
Summary: KloudMate Monitoring Agent
License: MIT
Group: Applications/System
BuildArch: x86_64
Source0: kmagent
Source1: postinst
Source2: kmagent.service
Source3: config.yaml


%description
Opentelemetry Collector for KloudMate platform

Requires: /bin/sh
Requires(post): systemd, systemctl
Requires(preun): systemd

%install
rm -rf %{buildroot}
# Create necessary directories in the build root
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_sysconfdir}
mkdir -p %{buildroot}%{_unitdir}
mkdir -p %{buildroot}%{scriptdir}
mkdir -p %{buildroot}%{confdir}

# Copy application binary - use Source0 notation
install -m 0755 %{SOURCE0} %{buildroot}%{_bindir}/%{appname}
install -m 0755 %{SOURCE1} %{buildroot}%{scriptdir}/postinst
install -m 0644 %{SOURCE2} %{buildroot}%{_unitdir}/%{appname}.service
install -m 0644 %{SOURCE3} %{buildroot}%{confdir}/config.yaml

# Copy configuration file (optional, could be done in postinst)
# install -m 0644 %{_sourcedir}/kmagent.conf %{buildroot}%{_sysconfdir}/%{appname}/kmagent.conf

# Copy the reusable (Debian-style) scripts
#install -m 0755 %{_sourcedir}/preinst  %{buildroot}%{scriptdir}/preinst
#install -m 0755 %{_sourcedir}/postinst %{buildroot}%{scriptdir}/postinst
#install -m 0755 %{_sourcedir}/prerm   %{buildroot}%{scriptdir}/prerm
#install -m 0755 %{_sourcedir}/postrm  %{buildroot}%{scriptdir}/postrm

# Example: Install systemd service file if you have one
# mkdir -p %{buildroot}%{_unitdir}
# install -m 0644 %{_sourcedir}/kmagent.service %{buildroot}%{_unitdir}/kmagent.service

# --- Scriptlet Sections ---
%pre
# Minimal pre-install actions if any, otherwise just exit
exit 0

%post
echo "Executing installed post-installation script %{scriptdir}/postinst ..."
if [ -x "%{scriptdir}/postinst" ]; then
    # Execute the script, passing the RPM argument ($1)
    "%{scriptdir}/postinst" "$1"
    SCRIPT_EXIT_CODE=$?
    if [ $SCRIPT_EXIT_CODE -ne 0 ]; then
         echo "ERROR: %{scriptdir}/postinst failed with exit code $SCRIPT_EXIT_CODE" >&2
         exit $SCRIPT_EXIT_CODE # <<< FAIL the installation if script fails
    fi
else
    echo "ERROR: %{scriptdir}/postinst not found or not executable. Installation cannot continue." >&2
    exit 1 # <<< FAIL the installation if script is missing
fi

# --- THEN, handle service start (if applicable and desired) ---
# This section only runs if the postinst script above succeeded.
if [ $1 -eq 1 ] ; then
    # Run only on initial installation (not upgrade)
    # Reload systemd to recognize the new service file
    /bin/systemctl daemon-reload >/dev/null 2>&1 || :
    # Enable the service to start on boot
    /bin/systemctl enable %{appname}.service >/dev/null 2>&1 || :
    # Start the service immediately (optional, could use 'enable --now')
    /bin/systemctl start %{appname}.service >/dev/null 2>&1 || :
fi
exit 0

%preun
if [ $1 -eq 0 ] ; then
    # Run only on final removal (not upgrade)
    # Stop the service
    /bin/systemctl stop %{appname}.service >/dev/null 2>&1 || :
    # Disable the service from starting on boot
    /bin/systemctl disable %{appname}.service >/dev/null 2>&1 || :
fi
exit 0

%postun
rm -rf %{confdir}
/bin/systemctl daemon-reload >/dev/null 2>&1 || :
exit 0

# --- File List ---
%files
%{_bindir}/%{appname}
%{_unitdir}/%{appname}.service
%{scriptdir}/postinst
%dir %{confdir}
%config(noreplace) %{confdir}/config.yaml

%changelog
# Add changelog entries here