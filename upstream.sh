#!/bin/sh

: ${SRCPKGDIR:=$HOME/projects/vanilla/srcpkgs}

# just ignore this!
Add_dependency()
{
	:
}

usage()
{
	echo "Usage: $(basename $0) [-s srcpkgdir] <all|pkgname>"
	exit 1
}

check_pkg()
{
	local pkgname="$1"

	if [ -r $SRCPKGDIR/${pkgname}/${pkgname}.template ]; then
		# skip subpkgs
		continue
	fi
	. $SRCPKGDIR/$pkgname/template
	watchver=$(./watch "$pkgname" 2>/dev/null)
	[ $? -ne 0 ] && continue
	[ -z "$watchver" -o "$watchver" = "" -o "$watchver" = "rcs" ] && continue
	echo "$pkgname: upstream version $watchver, srcpkgs $version."
}

while getopts "s:" opt; do
	case $opt in
		s) SRCPKGDIR="$OPTARG" ;;
		--) shift; break;;
	esac
done
shift $(($OPTIND - 1))

pkgname="$1"
[ $# -ne 1 ] && usage

if [ "$pkgname" = "all" ]; then
	for f in $(xbps-src list|awk '{print $1}'); do
		pkgname=$(xbps-uhelper getpkgname "$f")
		check_pkg "$pkgname"
	done
else
	check_pkg "$pkgname"
fi

exit 0
