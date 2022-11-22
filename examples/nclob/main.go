package main

import (
	"database/sql"
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

var buffer = `戸独稿案摯人三父載暮意豊回断合。川解山雑人神犯都会与福中国書兆約夢高。食過理処社本記殺宅惑名碁張島節誤東。批粉細入徴全鹿案易将基朝。同幡催速児載適歳療福実能説祝果権断起。絡素新著光番焦角阜細診技経退。数温交蓋能者確玉軍転嶋桜高全禁。西本宮本気育毎警利医費界。都売趣込定覧斎日兵氏惑座込給。同調香待速伝幸与健地真西亡。

込弱石精傾命実隆勤機復井歳落年主三向意。総報覚度校背社面演歓伏直学展法発塚患能。指朝上政入目紹文富韓再聖夜索八選裏聞告後。軍若間相始米業唱氷聞春水禁。傷最木希幕見球提性施辺扱無聞棋呼散逆渡。宅縄性断足久告転一秋終門邑目県投静学針万。問文転高陣準目質昨情更経激快投度秩応態新。追国軍系略第覧販試大春能。

配浮闘務償形成環断手成心新。提表権新図行祭海廃拡助記。座信個数向江新込省写街作横物。善図間理新都基馬競情影財法得将禁。北人経載謎者録頭組及捕表議想安漢規飯他型。捜進国約真軽京問家前朝東。引砲新女記日週能葬前太従義無東況作北値住。植動三催生視択推募約読供当国携分。危了面従記包坂浜北水与新見。説路無武外題食頂取詰食測朝禁。

号福定京主東今島度並棋具返出格細。尚澄総境勢報記医月去掲回待者。接住出影真害月申遠聞計文組義実。過型百興乗象要家生北点光意過西。偶要書団協由辞家図紙取新。回問奇第育快長批報教点社爆保木導。天島連間聞質分解禁奥部止務者禁原愛武。構国約多数情確担択治全報。輪購最者歩定覚利集細拡月変会。術挙動経防択男面後塾耳事待知部断相文。

界基強絵面会広歌政教華随。馬三及席幅幸際自孫男権齢意。内頼主成欠願成遠客新場力病日序各。富成利背点所詳述井問施段優務棄変投。芸態康庭秋他首場遊提医全奈。部過述円会目健南真修起定展駿。淡仲国用身少投規頭月史問禁期住炭狙。由組家叫川沢浮道併背都論家要治。会昭絡信全火賞全名体能指務将林。宅殊提稿挑確綱手優部三転間善。

裕購朗落持視望稿暮粉除害災天森心祭元。聞病百阜新完稿本向学督人惑原能碁悲住案経。得学人兆真目歌月天児者貼線乗南目属関。隆必放文置気延返伝安要士景松員載。様時明気能袴童湿末教遂飯動意松。外始豪的間肉坊会画勝載将。面稿会聞口告革細浅係面憲全色師院得。球編足記訓四中暮小固著戒際第。同要文索政賛紙件携開囲頼。

試転怒弾事点田門氏見秋在涯長元印。投分県陸対価茶立定転馬置連航。季条府賀自術融枝旅書泳那小地図。途神雄地士易詳稿阜大集県由。断回玄企択底四報包協遊検。弁稿条情科統満情車大写練視。選情季図倉勝生楽会王茨止度待新乳。止石久法新索結議形竹進応高福展旅。希拓麦制愛高加護三計週要跡車文口債。気西前真力動分就学特会要地断本細即専。

氏署鋭一図害本覚慮受図掲行権介士湧井他所。考解乗用自帯地教米軒及取貨面葉販。式条不校帯速速坂理詳菱南上。浜王魚引筑会芸負阪悪子球視。熊場押分大家止更課田井緊私山責地西折。発載然索成変山矢投遍年釣捕都静知可。医参専昔危証舞郎決墓国輪憶持造。扱別著視川破再調失汐止系活輪番無体。内線名能提界批生投江園知子読業渕信人長。

載由以部軽自復般役談喝得法段画見。楽黒崎就国女以使祖済片際。箕番求保止作米深棋条北蘇年済巡需議議脱提。張件毒宝寺化戦支沢送注位毎村野日選詳費。見断代芸海事容法事日狙煕峰両自報理作室。念景応相談経根坊出笑安極株強作料。入著変表分恵諸応同治者壮。逮言天系集係聞備権害務速記育。触告労表疑麻提団占育控天行写治猟名実朝。

領訪聞訪熱摘滅治基裏加兆忠収者自児提定層。職細削当官母学全田洋寄賞話蔵下優確。交出新際掲転公態測破喜性政製沖福子大害。齢残情合氷油提三高第勧浅悩外病城。神般園図藤務輸真馬並談総上新緒。表貸人著芸針転要望治水意願家問式祝人見氷。経人是魚会事浮義活常真追来写取質高点関郵。提三南禁律品度音違転言夜募。

器育活怪力足本難年川食伊申。少誕稿座方雇栃白済慎購泄調盛同。検法備察合固見付歳景聞衆容鷹断芸惑束。他覚陽加女掲回夫追騎柴苦趣聞流九進児。属的典法冬訪時今大録政注期保下滞郵刑闘似。入近新慢景対治身挑反捜難率質第最活示父。知場技戯道余休達動布速共。更百式覧極民題読動無経変州用営青前陸暮。跳権調案行無去告格涯済調良。

室観室裁止懸病問降世岡取第細。再物指挑直車個曲樹守世写提用挑理転灯泉表。外大山平理育残批運安夜考肉定応謙及陸美藤。売戦線納属著補言番著航方済容更無洋海。乗崎国談決最各図自院矢心断機。制信間待録芸投験幼際創終支証稿場野細。時照勝皇想摘用補更渡意大躍。旬故閣指図帳段打案身記化届世治無。藤利長身結傳晃稿首応政合国意投。

発因間接稿物詐瞳並攻試団。出存広沢計舎入選望比面力期甲井無康歩重恐。相以校富筑公売夜新化購故間仲応月日情。志脳老広明時覇賛真代芸次講返然。姫天盛攻断赤合科徳戒人年費応読。省格委念渡馬銃変対著田害幼山択。県中格覧図浦味量訴活開作社経。織途経更用混康開一能為署石前鉛満面者注応。滞科全遺慶長向字郎文王紙表。

来調氏大情断違半行柄安打売火再全載抽柳。区周歳松質現製能自庫治倒。能保然入浦雪安表必暑目自済授先。原住合品要保場受合名戸輪速祉児十。変研官太断倍巨暮学欠疾強弾辺済音必。範地方連種題樹時計単見併映葉隊連著夏取。臨了験将舞投断更上中多秀。教野点小故上登隣北職能重称公。演庁展身盗高探下催事座市質案欲線阪。

前統校方点芸行高条大注球在俳気。麗事室深覧当円高告社馬電日東界。定銀敬向打交一暗親更未蓮優年芸江理。道狼護裕指論稿聞戦寄法赤者業祉読。緯芽質根上暴殖軽図北読技特夫。記愛料入崎脱希掲兵区保覧月当下無介定情稿。事督済類守出野試控査市年終被赤質能。読情境首位感生団橋術経音東然乗犠。経配薬著埼伴助意書埼帯林戦天。

肉掲統倍北取立問職写午東情的線。広三情提受叟任海提州念幸。集表信士読国約料動展活査参油道無覧能。意耕英病名二理任用違活警官芸提禁森名。解用物規業昌建惑催身金垂。芸定外対折社供崎同報表特。置西富強追味録将超心会社史掲機。友覧捕止非成企平理去少載優大願風受。洞民木整事宙味網料堀手悩書正官。権締立成給渋城込権政算郎東常業持。

壊利療競探表甲夜堀新谷東康委言健日。入任継始設麻紀済多閣川用将経。再要速明林造惜必府志起友事棋点情特。都父依得応入日児強規速行不伎拍事就機舞女。追裕人押約和属天労文記権変投判野権県。千応当快跡故質試問顔中中任権投楽件年知。高明脅治竹乗両白子真情査営援省渡図。岩記償副救掲経匹圧集元能男。能倒真同問検株巨見空盟国。

特表続刑真府軽騒人市来殺護正真。安心織関庭変朝基見能組討迷著別投渋司年新。周私兄没写道体図後面界聞直性情出。毎候容問明情政個和少躍却通現拒。件育富体裁房施進葉相今最択場権庁発無遊会。宝記言感権初同初社載備苦初産突福史。需滋恵際扱大火画平震第原止。第避非小線論集本海持生載過気。報済属選美和用元毎界面実局真京故。

松者警明応勝早対載強済元者。止及分供屋土健王員物急換。詰院記木目真考塗堀罪筆界同古属事。決隊末権元康業以銃横主鈴。東決含件苦線卵設年本激掲進造標媛光第井。無合続予文蝶量本芸芸年仙回家速浜人判後死。箱必事解故会事選写決主費決。筆直柴吉近動用正活世野中集葉前会激人。業査軽所元進前自供大系力。界台広黒日火野掲神側解任披注訓。

記近学敬追過球長間通明優申円社違持果。次済行味挑間視済給格元活組今人食子歴。実属程秀無南示際郎出五主変側名備力税及。表孝現着活権申内嫌縮魂著。札著模案良権百前載件争婦治同同択配木郵。芸組主素力通女愛見勝表松声。携希件降別木価北険山利権意開降合面率時。朝体争泳面掲生写第手韓性能当両訳。山件尽権西作質在楽圧止中。`

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	VISIT_DATA  NCLOB
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	//val, err := ioutil.ReadFile("clob.json")
	//if err != nil {
	//	return err
	//}

	_, err := conn.Exec(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA) VALUES(1, :1)`, go_ora.NClob{String: buffer})
	if err != nil {
		return err
	}
	fmt.Println("1 row inserted: ", time.Now().Sub(t))
	return nil
}

func readWithSql(conn *sql.DB) error {
	t := time.Now()
	var (
		visitID   int64
		visitData go_ora.NVarChar
	)
	err := conn.QueryRow(`SELECT VISIT_ID, VISIT_DATA FROM GOORA_TEMP_VISIT`).Scan(&visitID, &visitData)
	if err != nil {
		return err
	}
	printLargeString("Data: ", string(visitData))
	fmt.Println("1 row read by sql: ", time.Now().Sub(t))
	return nil
}
func readWithOutPutPars(conn *sql.DB) error {
	t := time.Now()
	sqlText := `BEGIN
SELECT VISIT_DATA INTO :1 FROM GOORA_TEMP_VISIT WHERE VISIT_ID = 1;
END;`
	var data go_ora.NClob
	_, err := conn.Exec(sqlText, go_ora.Out{Dest: &data, Size: 100000})
	if err != nil {
		return err
	}
	printLargeString("Data: ", data.String)
	fmt.Println("1 row read by output parameters: ", time.Now().Sub(t))
	return nil
}
func printLargeString(prefix, data string) {
	if len(data) <= 25 {
		fmt.Println(prefix, data)
		return
	}
	temp := strings.ReplaceAll(data, "\r", "")
	temp = strings.ReplaceAll(temp, "\n", "\\n")
	fmt.Println(prefix, temp[:25], "...........", temp[len(temp)-25:], "\tsize: ", len(data))
}
func usage() {
	fmt.Println()
	fmt.Println("nclob")
	fmt.Println("  a code for using NClob by create table GOORA_TEMP_VISIT then insert then drop")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  nclob -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  nclob -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}
func main() {
	var server string
	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)

	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		fmt.Println("Can't open the driver", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection", err)
		return
	}

	err = createTable(conn)
	if err != nil {
		fmt.Println("Can't create table", err)
		return
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table", err)
		}
	}()

	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data: ", err)
		return
	}

	err = readWithSql(conn)
	if err != nil {
		fmt.Println("Can't read data with sql: ", err)
		return
	}

	err = readWithOutPutPars(conn)
	if err != nil {
		fmt.Println("Can't read data with output parameter: ", err)
		return
	}
}
