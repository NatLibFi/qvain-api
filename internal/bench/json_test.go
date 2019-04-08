package main

// ~/.local/gobin/easyjson -all -no_std_marshalers ../metax/metax.go
// ~/.local/gobin/ffjson -noencoder ../metax/metax.go

/*
	-rw-r--r-- 1 wouter wouter 14695 Nov 22 16:23 json_test.go

	BenchmarkJson-4                    10000            109502 ns/op [stdlib]
	BenchmarkJsonRoot-4                 5000            282306 ns/op [stdlib]
	BenchmarkFfjson-4                  10000            108967 ns/op [ffjson unmarshal]
	BenchmarkFfjsonRoot-4               5000            282358 ns/op [ffjson unmarshal]
	BenchmarkFfjson2-4                 10000            114120 ns/op [ffjson generated method]
	BenchmarkEasyJson-4                30000             57404 ns/op [easyjson generated method]
	BenchmarkJsoniter-4                30000             46464 ns/op [jsoniter unmarshal]
	BenchmarkJsoniterRoot-4            10000            206748 ns/op [jsoniter unmarshal]
*/

// -wvh- hmm... stdlib is faster than ffjson?

import (
	"testing"

	"encoding/json"
	"github.com/CSCfi/qvain-api/pkg/metax"
	//"github.com/mailru/easyjson"
	"github.com/json-iterator/go"
	"github.com/pquerna/ffjson/ffjson"
)

var jsonRecord string = `{
	"id":3,
	"alternate_record_set":["pid:urn:cr4"],
	"contract":{"id":1,"contract_json":{"quota":111204,"title":"Title of Contract 1","contact":[{"name":"Contact Name","email":"contact.email@csc.fi","phone":"+358501231234"}],"created":"2014-01-17T08:19:58Z","modified":"2014-01-17T08:19:58Z","validity":{"start_date":"2014-01-17T08:19:58Z"},"identifier":"optional:contract:identifier1","description":"Description of unknown length","organization":{"name":"Mysterious organization","organization_identifier":"1234567-1"},"related_service":[{"name":"Name of Service","identifier":"local:service:id"}]},"modified_by_api":"2017-05-15T13:07:22.559656","created_by_api":"2017-05-15T13:07:22.559656"},
	"data_catalog":{"id":1,"catalog_json":{"title":{"en":"Test data catalog name","fi":"Testidatakatalogin nimi"},"issued":"2014-02-27T08:19:58Z","homepage":[{"title":{"en":"Test website","fi":"Testi-verkkopalvelu"},"identifier":"http://testing.com"},{"title":{"en":"Another website","fi":"Toinen verkkopalvelu"},"identifier":"http://www.testing.fi"}],"language":[{"identifier":"http://lexvo.org/id/iso639-3/fin"},{"identifier":"http://lexvo.org/id/iso639-3/eng"}],"modified":"2014-01-17T08:19:58Z","harvested":false,"publisher":{"name":{"en":"Data catalog publisher organization","fi":"Datakatalogin julkaisijaorganisaatio"},"homepage":[{"title":{"en":"Publisher organization website","fi":"Julkaisijaorganisaation kotisivu"},"identifier":"http://www.publisher.fi/"}],"identifier":"http://isni.org/isni/0000000405129137"},"identifier":"pid:urn:catalog1","access_rights":{"type":[{"identifier":"http://purl.org/att/es/reference_data/access_type/access_type_open_access","pref_label":{"en":"Open","fi":"Avoin"}}],"license":[{"title":{"en":"CC BY 4.0","fi":"CC BY 4.0"},"identifier":"https://creativecommons.org/licenses/by/4.0/"}],"description":[{"fi":"Käyttöehtojen kuvaus"}],"has_rights_related_agent":[{"name":{"en":"A rights related organization","fi":"Oikeuksiin liittyvä organisaatio"},"identifier":"org_id"},{"name":{"en":"Org in ref data","fi":"Org referenssidatassa"},"email":"wahatever@madeupdomain.com","telephone":["+12353495823424"],"identifier":"http://purl.org/att/es/organization_data/organization/organization_10076"}]},"field_of_science":[{"identifier":"http://www.yso.fi/onto/okm-tieteenala/ta1172","pref_label":{"en":"Environmental sciences","fi":"Ympäristötiede"}}],"research_dataset_schema":"att"},"catalog_record_group_edit":"default-record-edit-group","catalog_record_group_create":"default-record-create-group","modified_by_api":"2017-05-15T13:07:22.559656","created_by_api":"2017-05-15T13:07:22.559656"},
	"research_dataset":{"files":[{"type":{"identifier":"http://purl.org/att/es/reference_data/file_type/file_type_text","pref_label":{"en":"Text","fi":"Teksti","und":"Teksti"}},"title":"File metadata title 6","identifier":"pid:urn:5","use_category":{"identifier":"http://purl.org/att/es/reference_data/use_category/use_category_source","pref_label":{"en":"Source material","fi":"Lähdeaineisto","und":"Lähdeaineisto"}}},{"type":{"identifier":"http://purl.org/att/es/reference_data/file_type/file_type_text","pref_label":{"en":"Text","fi":"Teksti","und":"Teksti"}},"title":"File metadata title 7","identifier":"pid:urn:dir:3","use_category":{"identifier":"http://purl.org/att/es/reference_data/use_category/use_category_source","pref_label":{"en":"Source material","fi":"Lähdeaineisto","und":"Lähdeaineisto"}}}],"title":{"en":"Wonderful Title"},"creator":[{"name":"Teppo Testaaja","@type":"Person","member_of":{"name":{"fi":"Mysteeriorganisaatio"},"@type":"Organization"}}],"curator":[{"name":"Rahikainen","@type":"Person","member_of":{"name":{"fi":"MysteeriOrganisaatio"},"@type":"Organization"},"identifier":"id:of:curator:rahikainen"}],"language":[{"title":{"aa":"English","af":"Engels","ak":"English","am":"እንግሊዝኛ","an":"Idioma anglés","ar":"لغة إنجليزية","as":"ইংৰাজী ভাষা","av":"Ингилис мацӀ","ay":"Inlish aru","az":"İngilis dili","ba":"Инглиз теле","be":"Англійская мова","bg":"Английски език","bm":"angilɛkan","bn":"ইংরেজি ভাষা","bo":"དབྱིན་ཇིའི་སྐད།","br":"saozneg","bs":"Engleski jezik","ca":"anglès","ce":"Ингалсан мотт","co":"Lingua inglese","cr":"ᐊᑲᔭᓯᒧᐃᐧᐣ","cs":"angličtina","cu":"Англїискъ ѩꙁꙑкъ","cv":"Акăлчан чĕлхи","cy":"Saesneg","da":"engelsk","de":"Englische Sprache","dv":"އިނގިރޭސި","dz":"ཨིང་ལིཤ་ཁ","ee":"Eŋlisigbe","el":"Αγγλική γλώσσα","en":"English language","eo":"Angla lingvo","es":"Idioma inglés","et":"Inglise keel","eu":"ingelesa","fa":"زبان انگلیسی","ff":"Engeleere","fi":"Englannin kieli","fo":"Enskt mál","fr":"anglais","fy":"Ingelsk","ga":"An Béarla","gd":"Beurla","gl":"Lingua inglesa","gn":"Inglyesñe&quot;ẽ","gu":"અંગ્રેજી ભાષા","gv":"Baarle","ha":"Turanci","he":"אנגלית","hi":"अंग्रेज़ी भाषा","hr":"Engleski jezik","ht":"Angle","hu":"Angol nyelv","hy":"Անգլերեն","ia":"Lingua anglese","id":"Bahasa Inggris","ie":"Angles","ig":"Asụsụ Inglish","ii":"ꑱꇩꉙ","io":"Angliana linguo","is":"enska","it":"Lingua inglese","iu":"ᖃᓪᓗᓈᑎᑐᑦ","ja":"英語","jv":"Basa Inggris","ka":"ინგლისური ენა","kg":"Kingelezi","ki":"Gĩthungũ","kk":"ағылшын тілі","kl":"tuluttut","km":"ភាសាអង់គ្លេស","kn":"ಇಂಗ್ಲೀಷ್","ko":"영어","ks":"اَنٛگیٖزۍ","ku":"Zimanê îngilîzî","kv":"Англия кыв","kw":"Sowsnek","ky":"Англис тили","la":"Lingua Anglica","lb":"Englesch","lg":"Lungereza","li":"Ingels","ln":"lingɛlɛ́sa","lo":"ພາສາອັງກິດ","lt":"Anglų kalba","lu":"Lingelesa","lv":"Angļu valoda","mg":"Fiteny anglisy","mi":"Reo Pākehā","mk":"Англиски јазик","ml":"ഇംഗ്ലീഷ്","mn":"Англи хэл","mr":"इंग्लिश भाषा","ms":"Bahasa Inggeris","mt":"Ingliż","my":"အင်္ဂလိပ်ဘာသာစကား","nb":"engelsk","nd":"isi-Ngisi","ne":"अङ्ग्रेजी भाषा","nl":"Engels","nn":"engelsk","no":"Engelsk","nv":"Bilagáana bizaad","ny":"Chingerezi","oc":"Anglés","om":"Ingliffa","or":"ଇଂରାଜୀ","os":"Англисаг æвзаг","pa":"ਅੰਗ੍ਰੇਜ਼ੀ ਭਾਸ਼ਾ","pi":"आंगलभाषा","pl":"Język angielski","ps":"انګرېزي ژبه","pt":"Língua inglesa","qu":"Inlish simi","rm":"Lingua englaisa","rn":"Icongereza","ro":"Limba engleză","ru":"Английский язык","rw":"Icyongereza","sa":"आङ्ग्लभाषा","sc":"Limba inglesa","se":"eaŋgalsgiella","sg":"Anglëe","sh":"Engleski jezik","si":"ඉංග්‍රීසි භාෂාව","sk":"angličtina","sl":"angleščina","sm":"Fa&quot;aperetania","sn":"Chirungu","so":"Ingiriisi","sq":"Gjuha angleze","sr":"Енглески језик","ss":"SíNgísi","st":"Senyesemane","su":"Basa Inggris","sv":"engelska","sw":"Kiingereza","ta":"ஆங்கிலம்","te":"ఆంగ్ల భాష","tg":"Забони англисӣ","th":"ภาษาอังกฤษ","ti":"እንግሊዝኛ","tk":"Iňlis dili","tl":"Wikang Ingles","tn":"Sekgoa","to":"lea fakapālangi","tr":"İngilizce","ts":"Xi Nghezi","tt":"Инглиз теле","tw":"English","ty":"Anglès","ug":"ئىنگىلىز تىلى","uk":"Англійська мова","ur":"انگریزی","uz":"Ingliz tili","vi":"Tiếng Anh","vo":"Linglänapük","wa":"Inglès","wo":"Wu-angalteer","xh":"isiNgesi","yi":"ענגליש","yo":"Èdè Gẹ̀ẹ́sì","za":"Vah Yinghgoz","zh":"英语","zu":"isiNgisi","ace":"Bahsa Inggréh","agq":"Kɨŋgele","aii":"ܠܫܢܐ ܐܢܓܠܝܐ","ang":"Nīƿu Englisc sprǣc","arz":"انجليزى","asa":"Kiingeredha","ast":"inglés","bar":"Englische Sproch","bas":"Hɔp u ŋgisì","bcl":"Ingles","bem":"Ichi Sungu","bez":"Hiingereza","bjn":"Bahasa Inggris","bpy":"ইংরেজি ঠার","brx":"अंग्रेज़ी","bug":"ᨅᨔ ᨕᨗᨋᨗᨔᨗ","byn":"እንግሊዝኛ","cdo":"Ĭng-ngṳ̄","ceb":"Iningles","cgg":"Orungyereza","chr":"ᎩᎵᏏ ᎦᏬᏂᎯᏍᏗ","ckb":"زمانی ئینگلیزی","cmn":"英文","crh":"İngliz tili","csb":"Anielsczi jãzëk","dav":"Kingereza","diq":"İngılızki","dje":"Inglisi senni","dsb":"Engelšćina","dyo":"angle","ebu":"Kĩthungu","ewo":"Ǹkɔ́bɔ éngəlís","ext":"Luenga ingresa","fil":"Ingles","frp":"Anglès","frr":"Ingelsk","fur":"Lenghe inglese","gag":"İngiliz dili","gan":"英語","got":"𐌰𐌲𐌲𐌹𐌻𐌰𐍂𐌰𐌶𐌳𐌰","gsw":"Änglisch","guz":"Kingereza","hak":"Yîn-ngî","haw":"‘Ōlelo Pelekania","hif":"English bhasa","hsb":"Jendźelšćina","ilo":"Pagsasao nga Ingglés","jbo":"glibau","jgo":"Aŋɡɛlúshi","jmc":"Kyingereza","kab":"Taglizit","kam":"Kingereza","kbd":"Инджылыбзэ","kde":"Chiingeleza","kea":"ingles","khq":"Inglisi senni","kkj":"yaman","kln":"kutitab Uingeresa","knn":"आंग्ल","koi":"Инглиш кыв","krc":"Ингилиз тил","ksb":"Kiingeeza","ksf":"riingɛrís","ksh":"Änglesch","lad":"Lingua inglesa","lag":"Kɨɨngeréesa","lbe":"Ингилис маз","lij":"Lèngoa ingleise","lmo":"Ingles","luo":"Kingereza","luy":"Lusungu","lzh":"L:英語","mas":"nkʉtʉ́k ɔ́ɔ̄ nkɨ́resa","mdf":"Англань кяль","mer":"Kĩngeretha","mfe":"angle","mgh":"Ingilishi","mhr":"Англичан йылме","mua":"zah Anglofoŋ","myv":"Англань кель","mzn":"اینگلیسی زبون","nah":"Inglatlahtōlli","nan":"Eng-gí","nap":"Lengua ngrese","naq":"Engels","nds":"Engelsche Spraak","new":"अंग्रेजी भाषा","nmg":"Ngɛ̄lɛ̄n","nnh":"ngilísè","nov":"Anglum","nso":"Seisimane","nus":"Thok liŋli̱thni","nyn":"Orungyereza","pap":"Ingles","pcd":"Inglé","pdc":"Englisch","pih":"Inglish","pms":"Lenga anglèisa","pnb":"انگریزی","rof":"Kiingereza","rue":"Анґліцькый язык","rup":"Limba anglicheascã","rwk":"Kyingereza","sah":"Аҥылычаанныы","saq":"Kingereza","sbp":"Ishingelesa","scn":"Lingua ngrisa","sco":"Inglis leid","seh":"inglês","ses":"Inglisi senni","shi":"ⵜⴰⵏⴳⵍⵉⵣⵜ","srn":"Ingristongo","stq":"Ängelske Sproake","swc":"Kingereza","swh":"Kiingereza","szl":"Angelsko godka","teo":"Kingereza","tig":"እንግሊዝኛ","tpi":"Tok Inglis","twq":"Inglisi senni","tzm":"Tanglizt","udm":"Англи кыл","und":"Englannin kieli","uzn":"Инглизча","vai":"ꕶꕱ","vec":"Łéngua inglexe","vro":"Inglüse kiil","vun":"Kyingereza","wae":"Engliš","wal":"እንግሊዝኛ","war":"Ininglis","wuu":"英语","xal":"Инглишин келн","xog":"Olungereza","yav":"íŋgilísé","yue":"英文","zea":"Iengels","zsm":"Inggeris","jv-x":"Basa Inggris","lt-x":"Onglu kalba","gsw-FR":"Englische Sprache","az-Cyrl":"инҝилисҹә","bs-Cyrl":"енглески","en-Dsrt":"𐐀𐑍𐑊𐐮𐑇","sr-Latn":"Engleski","cmn-Hant":"英文","shi-Latn":"tanglizt","uzn-Latn":"inglizcha","vai-Latn":"Poo","be-tarask":"Ангельская мова"},"identifier":"http://lexvo.org/id/iso639-3/eng"}],"modified":"2014-01-17T08:19:58Z","description":[{"en":"A descriptive description describing the contents of this dataset. Must be descriptive."}],"version_notes":["This version contains changes to x and y."],"urn_identifier":"pid:urn:cr3","total_byte_size":500,"preferred_identifier":"pid:urn:preferred:dataset4"},
	"preservation_state":3,
	"preservation_state_modified":"2017-11-22T08:47:12.572701",
	"mets_object_identifier":["a","b","c"],"dataset_group_edit":"default-dataset-edit-group",
	"modified_by_api":"2017-11-22T08:47:12.159962",
	"created_by_api":"2017-05-23T13:07:22.559656"
}`

func BenchmarkJson(b *testing.B) {
	rec := new(metax.MetaxRecord)
	for i := 0; i < b.N; i++ {
		err := json.Unmarshal([]byte(jsonRecord), &rec)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkJsonRoot(b *testing.B) {
	var top map[string]interface{}

	for i := 0; i < b.N; i++ {
		err := json.Unmarshal([]byte(jsonRecord), &top)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFfjson(b *testing.B) {
	rec := new(metax.MetaxRecord)
	for i := 0; i < b.N; i++ {
		err := ffjson.Unmarshal([]byte(jsonRecord), &rec)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFfjsonRoot(b *testing.B) {
	var top map[string]interface{}

	for i := 0; i < b.N; i++ {
		err := ffjson.Unmarshal([]byte(jsonRecord), &top)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFfjson2(b *testing.B) {
	rec := new(metax.MetaxRecord)
	for i := 0; i < b.N; i++ {
		err := rec.UnmarshalFFJSON([]byte(jsonRecord))
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkEasyJson(b *testing.B) {
	rec := new(metax.MetaxRecord)
	for i := 0; i < b.N; i++ {
		//err := rec.UnmarshalEasyJSON_EJ([]byte(jsonRecord))
		err := rec.UnmarshalJSON_EJ([]byte(jsonRecord))
		if err != nil {
			panic(err)
		}
	}
}

/*
func BenchmarkEasyJsonRoot(b *testing.B) {
	var top map[string]interface{}

	for i := 0; i < b.N; i++ {
		err := top.Unmarshal([]byte(jsonRecord))
		if err != nil {
			panic(err)
		}
	}
}
*/

func BenchmarkJsoniter(b *testing.B) {
	rec := new(metax.MetaxRecord)
	for i := 0; i < b.N; i++ {
		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsonRecord), &rec)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkJsoniterRoot(b *testing.B) {
	var top map[string]interface{}

	for i := 0; i < b.N; i++ {
		err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsonRecord), &top)
		if err != nil {
			panic(err)
		}
	}
}
