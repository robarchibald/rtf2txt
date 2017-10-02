package rtf2txt

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/EndFirstCorp/peekingReader"
)

// Text is used to convert an io.Reader containing RTF data into
// plain text
func Text(r io.Reader) (io.Reader, error) {
	buf := peekingReader.NewBufReader(r)

	var text bytes.Buffer
	var symbolStack stack
	for b, err := buf.ReadByte(); err == nil; b, err = buf.ReadByte() {
		switch b {
		case '\\':
			err := readControl(buf, &symbolStack, &text)
			if err != nil {
				return nil, err
			}
		case '{', '}':
		case '\n', '\r': // noop
		default:
			text.WriteByte(b)
		}
	}
	return &text, nil
}

func readControl(r peekingReader.Reader, s *stack, text *bytes.Buffer) error {
	data, err := peekingReader.ReadUntilAny(r, []byte{'\\', '{', '}', '\x00', '\t', '\f', ' ', '\n', '\r'})
	if err != nil {
		return err
	}
	control := string(data)
	if control == "*" { // this is an extended control sequence
		err := readUntilClosingBrace(r)
		return err
	}
	if isUnicode, u := getUnicode(control); isUnicode {
		text.WriteString(u)
		return nil
	}
	if control == "" {
		p, err := r.Peek(1)
		if err != nil {
			return err
		}
		if p[0] == '\\' || p[0] == '{' || p[0] == '}' { // this is an escaped character
			text.WriteByte(p[0])
			r.ReadByte()
			return nil
		}
		text.WriteByte('\n')
		return nil
	}
	err = handleBinary(r, control)
	if err != nil {
		return err
	}
	if symbol, found := convertSymbol(control); found {
		text.WriteString(symbol)
		return nil
	}

	ccontrol, _ := canonicalize(control)
	if isValue(ccontrol) || isToggle(ccontrol) || isDestination(control) { // Skip any parameter
		_, err := peekingReader.ReadUntilAny(r, []byte{'\\', '{', '}', '\n', '\r', ';'})
		if err != nil {
			return err
		}
		p, err := r.Peek(1)
		if err != nil {
			return err
		}
		if p[0] == ';' { // skip next if it is a semicolon
			r.ReadByte()
		}
	}
	s.Push(control)
	return nil
}

func getUnicode(control string) (bool, string) {
	if len(control) < 2 || control[0] != '\'' {
		return false, ""
	}

	var buf bytes.Buffer
	for i := 1; i < len(control); i++ {
		b := control[i]
		if b >= '0' && b <= '9' {
			buf.WriteByte(b)
		} else {
			break
		}
	}
	after := control[buf.Len()+1:]
	num, _ := strconv.ParseInt(buf.String(), 16, 16)
	return true, fmt.Sprintf("%c%s", num, after)
}

// canonicalize will return a control word with N in place of any digit
func canonicalize(control string) (string, int) {
	for i := 0; i < len(control); i++ {
		if control[i] >= '0' && control[i] <= '9' {
			num, err := strconv.Atoi(control[i:])
			if err != nil {
				return control, -1
			}
			return control[:i] + "N", num
		}
	}
	return control, -1
}

func handleBinary(r peekingReader.Reader, control string) error {
	if !strings.HasPrefix(control, "bin") || len(control) <= 3 { // wrong control type
		return nil
	}

	num, err := strconv.Atoi(control[3:])
	if err != nil { // not a number so skip
		return nil
	}
	_, err = r.ReadBytes(num)
	if err != nil {
		return err
	}
	return nil
}

func readText(r peekingReader.Reader) (string, error) {
	data, err := peekingReader.ReadUntilAny(r, []byte{'\\', '{', '}', '\n', '\r'})
	return string(data), err
}

func readUntilClosingBrace(r peekingReader.Reader) error {
	count := 1
	var b byte
	var err error
	for b, err = r.ReadByte(); err == nil; b, err = r.ReadByte() {
		switch b {
		case '{':
			count++
		case '}':
			count--
		}
		if count == 0 {
			return nil
		}
	}
	return err
}

func isDestination(symbol string) bool {
	switch symbol {
	case "aftncn", "aftnsep", "aftnsepc", "annotation", "atnauthor", "atndate", "atnicn", "atnid", "atnparent", "atnref", "atntime",
		"atrfend", "atrfstart", "author", "background", "bkmkend", "bkmkstart", "blipuid", "buptim", "category", "colorschememapping",
		"colortbl", "comment", "company", "creatim", "datafield ", "datastore", "defchp", "defpap", "do ", "doccomm", "docvar",
		"dptxbxtext", "ebcend", "ebcstart", "factoidname", "falt ", "fchars", "ffdeftext", "ffentrymcr", "ffexitmcr", "ffformat",
		"ffhelptext", "ffl", "ffname", "ffstattext", "field", "file ", "filetbl ", "fldinst", "fldrslt", "fname", "fontemb", "fontfile",
		"fonttbl", "footer", "footerf", "footerl", "footerr", "footnote", "formfield", "ftncn", "ftnsep", "ftnsepc", "g", "generator",
		"gridtbl", "header", "headerf", "headerl", "headerr", "hl", "hlfr", "hlinkbase", "hlloc", "hlsrc", "hsv", "htmltag", "info",
		"keycode", "keywords", "latentstyles", "lchars", "levelnumbers", "leveltext", "lfolevel", "linkval", "list", "listlevel",
		"listname", "listoverride", "listoverridetable", "listpicture", "liststylename", "listtable", "listtext", "lsdlockedexcept",
		"macc", "maccPr", "mailmerge", "maln", "malnScr", "manager", "margPr", "mbar", "mbarPr", "mbaseJc", "mbegChr", "mborderBox",
		"mborderBoxPr", "mbox", "mboxPr", "mchr", "mcount", "mctrlPr", "md", "mdeg", "mdegHide", "mden", "mdiff", "mdPr", "me",
		"mendChr", "meqArr", "meqArrPr", "mf", "mfName", "mfPr", "mfunc", "mfuncPr", "mgroupChr", "mgroupChrPr", "mgrow", "mhideBot",
		"mhideLeft", "mhideRight", "mhideTop", "mhtmltag", "mlim", "mlimloc", "mlimlow", "mlimlowPr", "mlimupp", "mlimuppPr", "mm",
		"mmaddfieldname", "mmath", "mmathPict", "mmathPr", "mmaxdist", "mmc", "mmcJc", "mmconnectstr", "mmconnectstrdata", "mmcPr",
		"mmcs", "mmdatasource", "mmheadersource", "mmmailsubject", "mmodso", "mmodsofilter", "mmodsofldmpdata", "mmodsomappedname",
		"mmodsoname", "mmodsorecipdata", "mmodsosort", "mmodsosrc ", "mmodsotable", "mmodsoudl", "mmodsoudldata", "mmodsouniquetag",
		"mmPr", "mmquery", "mmr", "mnary", "mnaryPr", "mnoBreak", "mnum", "mobjDist", "moMath", "moMathPara", "moMathParaPr", "mopEmu",
		"mphant", "mphantPr", "mplcHide", "mpos", "mr", "mrad", "mradPr", "mrPr", "msepChr", "mshow", "mshp", "msPre", "msPrePr", "msSub",
		"msSubPr", "msSubSup", "msSubSupPr", "msSup", "msSupPr", "mstrikeBLTR", "mstrikeH", "mstrikeTLBR", "mstrikeV", "msub", "msubHide",
		"msup", "msupHide", "mtransp", "mtype", "mvertJc", "mvfmf", "mvfml", "mvtof", "mvtol", "mzeroAsc", "mzeroDesc", "mzeroWid",
		"nesttableprops", "nextfile", "nonesttables", "objalias", "objclass", "objdata", "object", "objname", "objsect", "objtime",
		"oldcprops", "oldpprops", "oldsprops", "oldtprops", "oleclsid", "operator", "panose", "password", "passwordhash", "pgp", "pgptbl",
		"picprop", "pict", "pn ", "pntext ", "pntxta ", "pntxtb ", "printim", "private", "propname", "protend", "protstart", "protusertbl",
		"pxe", "result", "revtbl ", "revtim", "rsidtbl", "rtfN", "rxe", "shp", "shpgrp", "shpinst", "shppict", "shprslt", "shptxt", "sn",
		"sp", "staticval", "stylesheet", "subject", "sv", "svb", "tc", "template", "themedata", "title", "txe", "ud", "upr", "userprops",
		"wgrffmtfilter", "windowcaption", "writereservation", "writereservhash", "xe", "xform", "xmlattrname", "xmlattrvalue", "xmlclose",
		"xmlname", "xmlnstbl", "xmlopen", "fldtype":
		return true
	}
	return false
}

func isFlag(symbol string) bool {
	switch symbol {
	case "abslock", "additive", "adjustright", "aenddoc", "aendnotes", "afelev", "aftnbj", "aftnnalc", "aftnnar", "aftnnauc", "aftnnchi",
		"aftnnchosung", "aftnncnum", "aftnndbar", "aftnndbnum", "aftnndbnumd", "aftnndbnumk", "aftnndbnumt", "aftnnganada", "aftnngbnum",
		"aftnngbnumd", "aftnngbnumk", "aftnngbnuml", "aftnnrlc", "aftnnruc", "aftnnzodiac", "aftnnzodiacd", "aftnnzodiacl", "aftnrestart",
		"aftnrstcont", "aftntj", "allowfieldendsel", "allprot", "alntblind", "alt", "annotprot", "ansi", "ApplyBrkRules", "asianbrkrule",
		"autofmtoverride", "bdbfhdr", "bdrrlswsix", "bgbdiag", "bgcross", "bgdcross", "bgdkbdiag", "bgdkcross", "bgdkdcross", "bgdkfdiag",
		"bgdkhoriz", "bgdkvert", "bgfdiag", "bghoriz", "bgvert", "bkmkpub", "bookfold", "bookfoldrev", "box", "brdrb", "brdrbar", "brdrbtw",
		"brdrdash", "brdrdashd", "brdrdashdd", "brdrdashdot", "brdrdashdotdot", "brdrdashdotstr", "brdrdashsm", "brdrdb", "brdrdot", "brdremboss",
		"brdrengrave", "brdrframe", "brdrhair", "brdrinset", "brdrl", "brdrnil", "brdrnone", "brdroutset", "brdrr", "brdrs", "brdrsh",
		"brdrt", "brdrtbl", "brdrth", "brdrthtnlg", "brdrthtnmg", "brdrthtnsg", "brdrtnthlg", "brdrtnthmg", "brdrtnthsg", "brdrtnthtnlg",
		"brdrtnthtnmg", "brdrtnthtnsg", "brdrtriple", "brdrwavy", "brdrwavydb", "brkfrm", "bxe", "caccentfive", "caccentfour", "caccentone",
		"caccentsix", "caccentthree", "caccenttwo", "cachedcolbal", "cbackgroundone", "cbackgroundtwo", "cfollowedhyperlink", "chbgbdiag",
		"chbgcross", "chbgdcross", "chbgdkbdiag", "chbgdkcross", "chbgdkdcross", "chbgdkfdiag", "chbgdkhoriz", "chbgdkvert", "chbgfdiag",
		"chbghoriz", "chbgvert", "chbrdr", "chyperlink", "clbgbdiag", "clbgcross", "clbgdcross", "clbgdkbdiag", "clbgdkcross", "clbgdkdcross",
		"clbgdkfdiag", "clbgdkhor", "clbgdkvert", "clbgfdiag", "clbghoriz", "clbgvert", "clbrdrb", "clbrdrl", "clbrdrr", "clbrdrt", "cldel",
		"cldgll", "cldglu", "clFitText", "clhidemark", "clins", "clmgf", "clmrg", "clmrgd", "clmrgdr", "clNoWrap", "clshdrawnil", "clsplit",
		"clsplitr", "cltxbtlr", "cltxlrtb", "cltxlrtbv", "cltxtbrl", "cltxtbrlv", "clvertalb", "clvertalc", "clvertalt", "clvmgf", "clvmrg",
		"cmaindarkone", "cmaindarktwo", "cmainlightone", "cmainlighttwo", "collapsed", "contextualspace", "ctextone", "ctexttwo", "ctrl",
		"cvmme", "dbch", "defformat", "defshp", "dgmargin", "dgsnap", "dntblnsbdb", "dobxcolumn", "dobxmargin", "dobxpage", "dobymargin",
		"dobypage", "dobypara", "doctemp", "dolock", "donotshowcomments", "donotshowinsdel", "donotshowmarkup", "donotshowprops", "dpaendhol",
		"dpaendsol", "dparc", "dparcflipx", "dparcflipy", "dpastarthol", "dpastartsol", "dpcallout", "dpcoaccent", "dpcobestfit", "dpcoborder",
		"dpcodabs", "dpcodbottom", "dpcodcenter", "dpcodtop", "dpcominusx", "dpcominusy", "dpcosmarta", "dpcotdouble", "dpcotright",
		"dpcotsingle", "dpcottriple", "dpellipse", "dpendgroup", "dpfillbgpal", "dpfillfgpal", "dpgroup", "dpline", "dplinedado",
		"dplinedadodo", "dplinedash", "dplinedot", "dplinehollow", "dplinepal", "dplinesolid", "dppolygon", "dppolyline", "dprect",
		"dproundr", "dpshadow", "dptxbtlr", "dptxbx", "dptxlrtb", "dptxlrtbv", "dptxtbrl", "dptxtbrlv", "emfblip", "enddoc", "endnhere",
		"endnotes", "expshrtn", "faauto", "facenter", "facingp", "fafixed", "fahang", "faroman", "favar", "fbidi", "fbidis", "fbimajor",
		"fbiminor", "fdbmajor", "fdbminor", "fdecor", "felnbrelev", "fetch", "fhimajor", "fhiminor", "fjgothic", "fjminchou", "fldalt",
		"flddirty", "fldedit", "fldlock", "fldpriv", "flomajor", "flominor", "fmodern", "fnetwork", "fnil", "fnonfilesys", "forceupgrade",
		"formdisp", "formprot", "formshade", "fracwidth", "frmtxbtlr", "frmtxlrtb", "frmtxlrtbv", "frmtxtbrl", "frmtxtbrlv", "froman",
		"fromtext", "fscript", "fswiss", "ftech", "ftnalt", "ftnbj", "ftnil", "ftnlytwnine", "ftnnalc", "ftnnar", "ftnnauc", "ftnnchi",
		"ftnnchosung", "ftnncnum", "ftnndbar", "ftnndbnum", "ftnndbnumd", "ftnndbnumk", "ftnndbnumt", "ftnnganada", "ftnngbnum", "ftnngbnumd",
		"ftnngbnumk", "ftnngbnuml", "ftnnrlc", "ftnnruc", "ftnnzodiac", "ftnnzodiacd", "ftnnzodiacl", "ftnrestart", "ftnrstcont", "ftnrstpg",
		"ftntj", "fttruetype", "fvaliddos", "fvalidhpfs", "fvalidmac", "fvalidntfs", "gutterprl", "hich", "horzdoc", "horzsect", "hrule",
		"htmautsp", "htmlbase", "hwelev", "indmirror", "indrlsweleven", "intbl", "ixe", "jclisttab", "jcompress", "jexpand", "jis",
		"jpegblip", "jsksu", "keep", "keepn", "krnprsnet", "landscape", "lastrow", "levelpicturenosize", "linebetcol", "linecont", "lineppage",
		"linerestart", "linkself", "linkstyles", "listhybrid", "listoverridestartat", "lnbrkrule", "lndscpsxn", "lnongrid", "loch", "ltrch",
		"ltrdoc", "ltrpar", "ltrrow", "ltrsect", "lvltentative", "lytcalctblwd", "lytexcttp", "lytprtmet", "lyttblrtgr", "mac", "macpict",
		"makebackup", "margmirror", "margmirsxn", "mmattach", "mmblanklines", "mmdatatypeaccess", "mmdatatypeexcel", "mmdatatypefile",
		"mmdatatypeodbc", "mmdatatypeodso", "mmdatatypeqt", "mmdefaultsql", "mmdestemail", "mmdestfax", "mmdestnewdoc", "mmdestprinter",
		"mmlinktoquery", "mmmaintypecatalog", "mmmaintypeemail", "mmmaintypeenvelopes", "mmmaintypefax", "mmmaintypelabels",
		"mmmaintypeletters", "mmshowdata", "msmcap", "muser", "mvf", "mvt", "newtblstyruls", "noafcnsttbl", "nobrkwrptbl", "nocolbal",
		"nocompatoptions", "nocwrap", "nocxsptable", "noextrasprl", "nofeaturethrottle", "nogrowautofit", "noindnmbrts", "nojkernpunct",
		"nolead", "noline", "nolnhtadjtbl", "nonshppict", "nooverflow", "noproof", "noqfpromote", "nosectexpand", "nosnaplinegrid",
		"nospaceforul", "nosupersub", "notabind", "notbrkcnstfrctbl", "notcvasp", "notvatxbx", "nouicompat", "noultrlspc", "nowidctlpar",
		"nowrap", "nowwrap", "noxlattoyen", "objattph", "objautlink", "objemb", "objhtml", "objicemb", "objlink", "objlock", "objocx",
		"objpub", "objsetsize", "objsub", "objupdate", "oldas", "oldlinewrap", "otblrul", "overlay", "pagebb", "pard", "pc", "pca",
		"pgbrdrb", "pgbrdrfoot", "pgbrdrhead", "pgbrdrl", "pgbrdrr", "pgbrdrsnap", "pgbrdrt", "pgnbidia", "pgnbidib", "pgnchosung",
		"pgncnum", "pgncont", "pgndbnum", "pgndbnumd", "pgndbnumk", "pgndbnumt", "pgndec", "pgndecd", "pgnganada", "pgngbnum",
		"pgngbnumd", "pgngbnumk", "pgngbnuml", "pgnhindia", "pgnhindib", "pgnhindic", "pgnhindid", "pgnhnsc", "pgnhnsh", "pgnhnsm",
		"pgnhnsn", "pgnhnsp", "pgnid", "pgnlcltr", "pgnlcrm", "pgnrestart", "pgnthaia", "pgnthaib", "pgnthaic", "pgnucltr", "pgnucrm",
		"pgnvieta", "pgnzodiac", "pgnzodiacd", "pgnzodiacl", "phcol", "phmrg", "phpg", "picbmp", "picscaled", "pindtabqc", "pindtabql",
		"pindtabqr", "plain", "pmartabqc", "pmartabql", "pmartabqr", "pnacross", "pnaiu", "pnaiud", "pnaiueo", "pnaiueod", "pnbidia", "pnbidib",
		"pncard", "pnchosung", "pncnum", "pndbnum", "pndbnumd", "pndbnumk", "pndbnuml", "pndbnumt", "pndec", "pndecd", "pnganada", "pngblip",
		"pngbnum", "pngbnumd", "pngbnumk", "pngbnuml", "pnhang", "pniroha", "pnirohad", "pnlcltr", "pnlcrm", "pnlvlblt", "pnlvlbody",
		"pnlvlcont", "pnnumonce", "pnord", "pnordt", "pnprev", "pnqc", "pnql", "pnqr", "pnrestart", "pnrnot", "pnucltr", "pnucrm", "pnuld",
		"pnuldash", "pnuldashd", "pnuldashdd", "pnuldb", "pnulhair", "pnulnone", "pnulth", "pnulw", "pnulwave", "pnzodiac", "pnzodiacd",
		"pnzodiacl", "posxc", "posxi", "posxl", "posxo", "posxr", "posyb", "posyc", "posyil", "posyin", "posyout", "posyt", "prcolbl",
		"printdata", "psover", "ptabldot", "ptablmdot", "ptablminus", "ptablnone", "ptabluscore", "pubauto", "pvmrg", "pvpara", "pvpg", "qc",
		"qd", "qj", "ql", "qr", "qt", "rawclbgbdiag", "rawclbgcross", "rawclbgdcross", "rawclbgdkbdiag", "rawclbgdkcross", "rawclbgdkdcross",
		"rawclbgdkfdiag", "rawclbgdkhor", "rawclbgdkvert", "rawclbgfdiag", "rawclbghoriz", "rawclbgvert", "readonlyrecommended", "readprot",
		"remdttm", "rempersonalinfo", "revisions", "revprot", "rsltbmp", "rslthtml", "rsltmerge", "rsltpict", "rsltrtf", "rslttxt", "rtlch",
		"rtldoc", "rtlgutter", "rtlpar", "rtlrow", "rtlsect", "saftnnalc", "saftnnar", "saftnnauc", "saftnnchi", "saftnnchosung", "saftnncnum",
		"saftnndbar", "saftnndbnum", "saftnndbnumd", "saftnndbnumk", "saftnndbnumt", "saftnnganada", "saftnngbnum", "saftnngbnumd",
		"saftnngbnumk", "saftnngbnuml", "saftnnrlc", "saftnnruc", "saftnnzodiac", "saftnnzodiacd", "saftnnzodiacl", "saftnrestart",
		"saftnrstcont", "sautoupd", "saveinvalidxml", "saveprevpict", "sbkcol", "sbkeven", "sbknone", "sbkodd", "sbkpage", "sbys",
		"scompose", "sectd", "sectdefaultcl", "sectspecifycl", "sectspecifygenN", "sectspecifyl", "sectunlocked", "sftnbj", "sftnnalc",
		"sftnnar", "sftnnauc", "sftnnchi", "sftnnchosung", "sftnncnum", "sftnndbar", "sftnndbnum", "sftnndbnumd", "sftnndbnumk", "sftnndbnumt",
		"sftnnganada", "sftnngbnum", "sftnngbnumd", "sftnngbnumk", "sftnngbnuml", "sftnnrlc", "sftnnruc", "sftnnzodiac", "sftnnzodiacd",
		"sftnnzodiacl", "sftnrestart", "sftnrstcont", "sftnrstpg", "sftntj", "shidden", "shift", "shpbxcolumn", "shpbxignore", "shpbxmargin",
		"shpbxpage", "shpbyignore", "shpbymargin", "shpbypage", "shpbypara", "shplockanchor", "slocked", "snaptogridincell", "softcol",
		"softline", "softpage", "spersonal", "spltpgpar", "splytwnine", "sprsbsp", "sprslnsp", "sprsspbf", "sprstsm", "sprstsp", "spv",
		"sqformat", "sreply", "stylelock", "stylelockbackcomp", "stylelockenforced", "stylelockqfset", "stylelocktheme", "sub",
		"subfontbysize", "super", "swpbdr", "tabsnoovrlp", "taprtl", "tbllkbestfit", "tbllkborder", "tbllkcolor", "tbllkfont", "tbllkhdrcols",
		"tbllkhdrrows", "tbllklastcol", "tbllklastrow", "tbllknocolband", "tbllknorowband", "tbllkshading", "tcelld", "tcn", "titlepg",
		"tldot", "tleq", "tlhyph", "tlmdot", "tlth", "tlul", "toplinepunct", "tphcol", "tphmrg", "tphpg", "tposxc", "tposxi", "tposxl",
		"tposxo", "tposxr", "tposyb", "tposyc", "tposyil", "tposyin", "tposyout", "tposyt", "tpvmrg", "tpvpara", "tpvpg", "tqc", "tqdec",
		"tqr", "transmf", "trbgbdiag", "trbgcross", "trbgdcross", "trbgdkbdiag", "trbgdkcross", "trbgdkdcross", "trbgdkfdiag", "trbgdkhor",
		"trbgdkvert", "trbgfdiag", "trbghoriz", "trbgvert", "trbrdrb", "trbrdrh", "trbrdrl", "trbrdrr", "trbrdrt", "trbrdrv", "trhdr",
		"trkeep", "trkeepfollow", "trowd", "trqc", "trql", "trqr", "truncatefontheight", "truncex", "tsbgbdiag", "tsbgcross", "tsbgdcross",
		"tsbgdkbdiag", "tsbgdkcross", "tsbgdkdcross", "tsbgdkfdiag", "tsbgdkhor", "tsbgdkvert", "tsbgfdiag", "tsbghoriz", "tsbgvert",
		"tsbrdrb", "tsbrdrdgl", "tsbrdrdgr", "tsbrdrh", "tsbrdrl", "tsbrdrr", "tsbrdrt", "tsbrdrv", "tscbandhorzeven",
		"tscbandhorzodd", "tscbandverteven", "tscbandvertodd", "tscfirstcol", "tscfirstrow", "tsclastcol", "tsclastrow", "tscnecell",
		"tscnwcell", "tscsecell", "tscswcell", "tsd", "tsnowrap", "tsrowd", "tsvertalb", "tsvertalc", "tsvertalt", "twoonone",
		"txbxtwalways", "txbxtwfirst", "txbxtwfirstlast", "txbxtwlast", "txbxtwno", "uld", "ulnone", "ulw", "useltbaln", "usenormstyforlist",
		"usexform", "utinl", "vertal", "vertalb", "vertalc", "vertalj", "vertalt", "vertdoc", "vertsect", "viewnobound", "webhidden",
		"widctlpar", "widowctrl", "wpjst", "wpsp", "wraparound", "wrapdefault", "wrapthrough", "wraptight", "wraptrsp", "wrppunct",
		"xmlattr", "xmlsdttcell", "xmlsdttpara", "xmlsdttregular", "xmlsdttrow", "xmlsdttunknown", "yxe", "mlit", "mmfttypeaddress",
		"mmfttypebarcode", "mmfttypedbcolumn", "mmfttypemapped", "mmfttypenull", "mmfttypesalutation", "mnor", "date", "time", "wpeqn":
		return true
	}
	return false
}

func isToggle(symbol string) bool {
	switch symbol {
	case "ab", "absnoovrlpN", "acaps", "acccircle", "acccomma", "accdot", "accnone", "accunderdot", "ai", "aoutl", "ascaps", "ashad",
		"aspalpha", "aspnum", "astrike", "aul", "auld", "auldb", "aulnone", "aulw", "b", "caps", "deleted", "disabled", "embo", "htmlrtf",
		"hyphauto", "hyphcaps", "hyphpar", "i", "impr", "outl", "pnb", "pncaps", "pni", "pnscaps", "pnstrike", "pnul", "protect", "revised",
		"saautoN", "sbautoN", "scaps", "shad", "strike", "striked1", "trautofitN", "ul", "uldash", "uldashd", "uldashdd", "uldb", "ulhair",
		"ulhwave", "ulldash", "ulth", "ulthd", "ulthdash", "ulthdashd", "ulthdashdd", "ulthldash", "ululdbwave", "ulwave", "v":
		return true
	}
	return false
}

func isValue(symbol string) bool {
	switch symbol {
	case "abshN", "abswN", "acfN", "adeffN", "adeflangN", "adnN", "aexpndN", "afN", "afsN", "aftnstartN", "alangN", "animtextN",
		"ansicpgN", "aupN", "binfsxnN", "binN", "binsxnN", "bkmkcolfN", "bkmkcollN", "bliptagN", "blipupiN", "blueN", "bookfoldsheetsN",
		"brdrartN", "brdrcfN", "brdrwN", "brspN", "cbN", "cbpatN", "cchsN", "cellxN", "cfN", "cfpatN", "cgridN", "charrsidN", "charscalexN",
		"chcbpatN", "chcfpatN", "chhresN", "chshdngN", "clcbpatN", "clcbpatrawN", "clcfpatN", "clcfpatrawN", "cldelauthN", "cldeldttmN",
		"clftsWidthN", "clinsauthN", "clinsdttmN", "clmrgdauthN", "clmrgddttmN", "clpadbN", "clpadfbN", "clpadflN", "clpadfrN", "clpadftN",
		"clpadlN", "clpadrN", "clpadtN", "clshdngN", "clshdngrawN", "clspbN", "clspfbN", "clspflN", "clspfrN", "clspftN", "clsplN",
		"clsprN", "clsptN", "clwWidthN", "colnoN", "colsN", "colsrN", "colsxN", "colwN", "cpgN", "crauthN", "crdateN", "cshadeN", "csN",
		"ctintN", "ctsN", "cufiN", "culiN", "curiN", "deffN", "deflangfeN", "deflangN", "deftabN", "delrsidN", "dfrauthN", "dfrdateN",
		"dfrmtxtxN", "dfrmtxtyN", "dfrstart", "dfrstop", "dfrxst", "dghoriginN", "dghshowN", "dghspaceN", "dgvoriginN", "dgvshowN",
		"dgvspaceN", "dibitmapN", "dnN", "doctypeN", "dodhgtN", "donotembedlingdataN", "donotembedsysfontN", "dpaendlN", "dpaendwN",
		"dpastartlN", "dpastartwN", "dpcoaN", "dpcodescentN", "dpcolengthN", "dpcooffsetN", "dpcountN", "dpfillbgcbN", "dpfillbgcgN",
		"dpfillbgcrN", "dpfillbggrayN", "dpfillfgcbN", "dpfillfgcgN", "dpfillfgcrN", "dpfillfggrayN", "dpfillpatN", "dplinecobN",
		"dplinecogN", "dplinecorN", "dplinegrayN", "dplinewN", "dppolycountN", "dpptxN", "dpptyN", "dpshadxN", "dpshadyN", "dptxbxmarN",
		"dpxN", "dpxsizeN", "dpyN", "dpysizeN", "dropcapliN", "dropcaptN", "dsN", "dxfrtextN", "dyN", "edminsN", "enforceprotN", "expndN",
		"expndtwN", "fbiasN", "fcharsetN", "fcsN", "fetN", "ffdefresN", "ffhaslistboxN", "ffhpsN", "ffmaxlenN", "ffownhelpN", "ffownstatN",
		"ffprotN", "ffrecalcN", "ffresN", "ffsizeN", "fftypeN", "fftypetxtN", "fidN", "fiN", "fittextN", "fN", "fnN", "footeryN", "fosnumN",
		"fprqN", "frelativeN", "fromhtmlN", "fsN", "ftnstartN", "gcwN", "greenN", "grfdoceventsN", "gutterN", "guttersxnN", "headeryN",
		"highlightN", "horzvertN", "hresN", "hrN", "hyphconsecN", "hyphhotzN", "idN", "ignoremixedcontentN", "ilfomacatclnupN", "ilvlN",
		"insrsidN", "ipgpN", "irowbandN", "irowN", "itapN", "kerningN", "ksulangN", "langfeN", "langfenpN", "langN", "langnpN", "lbrN",
		"levelfollowN", "levelindentN", "leveljcN", "leveljcnN", "levellegalN", "levelN", "levelnfcN", "levelnfcnN", "levelnorestartN",
		"leveloldN", "levelpictureN", "levelprevN", "levelprevspaceN", "levelspaceN", "levelstartatN", "leveltemplateidN", "liN", "linemodN",
		"linestartN", "linestartsN", "linexN", "linN", "lisaN", "lisbN", "listidN", "listoverridecountN", "listoverrideformatN",
		"listrestarthdnN", "listsimpleN", "liststyleidN", "listtemplateidN", "lsdlockeddefN", "lsdlockedN", "lsdprioritydefN", "lsdpriorityN",
		"lsdqformatdefN", "lsdqformatN", "lsdsemihiddendefN", "lsdsemihiddenN", "lsdstimaxN", "lsdunhideuseddefN", "lsdunhideusedN",
		"lsN", "margbN", "margbsxnN", "marglN", "marglsxnN", "margrN", "margrsxnN", "margSzN", "margtN", "margtsxnN", "mbrkBinN",
		"mbrkBinSubN", "mbrkN", "mcGpN", "mcGpRuleN", "mcSpN", "mdefJcN", "mdiffStyN", "mdispdefN", "minN", "minterSpN", "mintLimN",
		"mintraSpN", "mjcN", "mlMarginN", "mmathFontN", "mmerrorsN", "mmjdsotypeN", "mmodsoactiveN", "mmodsocoldelimN", "mmodsocolumnN",
		"mmodsodynaddrN", "mmodsofhdrN", "mmodsofmcolumnN", "mmodsohashN", "mmodsolidN", "mmreccurN", "mnaryLimN", "moN", "mpostSpN",
		"mpreSpN", "mrMarginN", "mrSpN", "mrSpRuleN", "mscrN", "msmallFracN", "mstyN", "mvauthN", "mvdateN", "mwrapIndentN", "mwrapRightN",
		"nofcharsN", "nofcharswsN", "nofpagesN", "nofwordsN", "objalignN", "objcropbN", "objcroplN", "objcroprN", "objcroptN", "objhN",
		"objscalexN", "objscaleyN", "objtransyN", "objwN", "ogutterN", "outlinelevelN", "paperhN", "paperwN", "pararsidN", "pgbrdroptN",
		"pghsxnN", "pgnhnN", "pgnstartN", "pgnstartsN", "pgnxN", "pgnyN", "pgwsxnN", "picbppN", "piccropbN", "piccroplN", "piccroprN",
		"piccroptN", "pichgoalN", "pichN", "picscalexN", "picscaleyN", "picwgoalN", "picwN", "pmmetafileN", "pncfN", "pnfN", "pnfsN",
		"pnindentN", "pnlvlN", "pnrauthN", "pnrdateN", "pnrnfcN", "pnrpnbrN", "pnrrgbN", "pnrstartN", "pnrstopN", "pnrxstN", "pnspN",
		"pnstartN", "posnegxN", "posnegyN", "posxN", "posyN", "prauthN", "prdateN", "proptypeN", "protlevelN", "pszN", "pwdN", "qkN", "redN",
		"relyonvmlN", "revauthdelN", "revauthN", "revbarN", "revdttmdelN", "revdttmN", "revpropN", "riN", "rinN", "rsidN", "rsidrootN",
		"saftnstartN", "saN", "sbasedonN", "sbN", "secN", "sectexpandN", "sectlinegridN", "sectrsidN", "sftnstartN", "shadingN",
		"showplaceholdtextN", "showxmlerrorsN", "shpbottomN", "shpfblwtxtN", "shpfhdrN", "shpleftN", "shplidN", "shprightN", "shptopN",
		"shpwrkN", "shpwrN", "shpzN", "slinkN", "slmultN", "slN", "sN", "snextN", "softlheightN", "spriorityN", "srauthN", "srdateN",
		"ssemihiddenN", "stextflowN", "stshfbiN", "stshfdbchN", "stshfhichN", "stshflochN", "stylesortmethodN", "styrsidN", "subdocumentN",
		"sunhideusedN", "tblindN", "tblindtypeN", "tblrsidN", "tbN", "tcfN", "tclN", "tdfrmtxtBottomN", "tdfrmtxtLeftN", "tdfrmtxtRightN",
		"tdfrmtxtTopN", "themelangcsN", "themelangfeN", "themelangN", "tposnegxN", "tposnegyN", "tposxN", "tposyN", "trackformattingN",
		"trackmovesN", "trauthN", "trcbpatN", "trcfpatN", "trdateN", "trftsWidthAN", "trftsWidthBN", "trftsWidthN", "trgaphN", "trleftN",
		"trpaddbN", "trpaddfbN", "trpaddflN", "trpaddfrN", "trpaddftN", "trpaddlN", "trpaddrN", "trpaddtN", "trpadobN", "trpadofbN",
		"trpadoflN", "trpadofrN", "trpadoftN", "trpadolN", "trpadorN", "trpadotN", "trpatN", "trrhN", "trshdngN", "trspdbN", "trspdfbN",
		"trspdflN", "trspdfrN", "trspdftN", "trspdlN", "trspdrN", "trspdtN", "trspobN", "trspofbN", "trspoflN", "trspofrN", "trspoftN",
		"trspolN", "trsporN", "trspotN", "trwWidthAN", "trwWidthBN", "trwWidthN", "tscbandshN", "tscbandsvN", "tscellcbpatN", "tscellcfpatN",
		"tscellpaddbN", "tscellpaddfbN", "tscellpaddflN", "tscellpaddfrN", "tscellpaddftN", "tscellpaddlN", "tscellpaddrN", "tscellpaddtN",
		"tscellpctN", "tscellwidthftsN", "tscellwidthN", "tsN", "twoinoneN", "txN", "ucN", "ulcN", "uN", "upN", "urtfN", "validatexmlN",
		"vernN", "versionN", "viewbkspN", "viewkindN", "viewscaleN", "viewzkN", "wbitmapN", "wbmbitspixelN", "wbmplanesN", "wbmwidthbyteN",
		"wmetafileN", "xefN", "xmlattrnsN", "xmlnsN", "yrN", "ytsN":
		return true
	}
	return false
}

func convertSymbol(symbol string) (string, bool) {
	switch symbol {
	case "bullet":
		return "*", true
	case "chdate", "chdpa", "chdpl":
		return time.Now().Format("2005-01-02"), true
	case "chtime":
		return time.Now().Format("4:56 pm"), true
	case "emdash", "endash":
		return "-", true
	case "lquote", "rquote":
		return "'", true
	case "ldblquote", "rdblquote":
		return "\"", true
	case "line", "lbrN":
		return "\n", true
	case "cell", "column", "emspace", "enspace", "qmspace", "nestcell", "nestrow", "page", "par", "row", "sect", "tab":
		return " ", true
	case "|", "~", "-", "_", ":":
		return symbol, true
	case "chatn", "chftn", "chftnsep", "chftnsepc", "chpgn", "sectnum", "ltrmark", "rtlmark", "zwbo", "zwj", "zwnbo", "zwnj", "softcol",
		"softline", "softpage":
		return "", true
	default:
		return "", false
	}
}
