# Build the PAPI Go package.
# By Scott Pakin <pakin@lanl.gov>

include $(GOROOT)/src/Make.inc

VERSION=1.0

TARG=papi

CGOFILES=\
	papi.go\
	papi-high.go\
	papi-low.go\
	papi-mh.go\
	papi-errno.go\
	papi-event.go\
	papi-emod.go\

DISTFILES=\
	papi.go\
	papi-high.go\
	papi-low.go\
	papi-mh.go\
	consts2code\
	Makefile\
	papi_test.go\
	papi_hl_test.go\
	papi_ll_test.go\

include $(GOROOT)/src/Make.pkg

# ---------------------------------------------------------------------------

# We use a helper Perl script, consts2code, to generate papi-errno.go,
# papi-event.go, and papi-emod.go.

PERL=perl
AWK=awk
PAPI_INCDIR:=$(dir $(shell $(PERL) consts2code papi.h))

papi-errno.go: consts2code $(PAPI_INCDIR)/papi.h
	$(PERL) consts2code \
	    papi.h \
	    --format='%s os.Error = Errno(C.PAPI_%s)' \
	    --comment="The following constants can be returned as Errno values from PAPI functions." \
	    --keep='#define' \
	    --keep='PAPI_E.*-\d' | \
	  $(AWK) '{print} /import/ {print "import \"os\""}' | \
	  sed 's/const /var /' > papi-errno.go

papi-event.go: consts2code $(PAPI_INCDIR)/papiStdEventDefs.h
	$(PERL) consts2code \
	    papiStdEventDefs.h \
	    --format='%s Event = C.PAPI_%s' \
	    --comment="The following constants represent PAPI's standard event types." \
	    --keep='#define' \
	    --keep='_idx' | grep -v PAPI_END > papi-event.go

papi-emod.go: consts2code $(PAPI_INCDIR)/papi.h
	$(PERL) consts2code \
	    papi.h \
	    --format='%s EventModifier = C.PAPI_%s' \
	    --comment="An EventModifier filters the set of events returned by EnumEvents()." \
	    --keep='PAPI_(NTV|PRESET_BIT|ENUM|PRESET_ENUM_AVAIL)' \
	    --no-ifdef > papi-emod.tmp1
	echo "" > papi-emod.tmp2
	echo "// Map each of the above to a string." >> papi-emod.tmp2
	$(PERL) -e 'while (<>) {/(PRESET_BIT_\w+)\s+EventModifier =/ && push @sym, $$1} printf "var presetBitToString=map[EventModifier]string{\n%s}\n", join ",\n", map {"$$_:\"PAPI_$$_\""} @sym' papi-emod.tmp1 >> papi-emod.tmp2
	cat papi-emod.tmp1 papi-emod.tmp2 | gofmt > papi-emod.go
	$(RM) papi-emod.tmp1 papi-emod.tmp2

CLEANFILES += papi-errno.go papi-event.go papi-emod.go papi-emod.tmp1 papi-emod.tmp2

# ---------------------------------------------------------------------------

FULLNAME=gopapi-$(VERSION)

dist: $(DISTFILES)
	mkdir $(FULLNAME)
	cp $(DISTFILES) $(FULLNAME)
	tar -czf $(FULLNAME).tar.gz $(FULLNAME)
	$(RM) -r $(FULLNAME)
	tar -tzvf $(FULLNAME).tar.gz
