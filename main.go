package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/fatih/color"
)

// Json dosyamızdaki değerleri structa dönüştürmek için kullandığımız yapı.
// Basit bir şekilde dönüştürmek için: https://mholt.github.io/json-to-go/
type Video struct {
	Items []struct {
		Statistics struct {
			ViewCount    string `json:"viewCount"`
			LikeCount    string `json:"likeCount"`
			DislikeCount string `json:"dislikeCount"`
			CommentCount string `json:"commentCount"`
		} `json:"statistics"`
	} `json:"items"`
}

// program boyunca kullanacağımız renk değişkenleri
var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

func main() {
	urlFlag := flag.String("url", "", "Youtube video istatistiğini almak için kullandığımız flag")
	apiFlag := flag.String("api", "", "Bütün işlemleri yaparken kullanacğımız API Keyi")
	// Flaglarimizi parse etmek için bunu kullanmak zorundayız
	flag.Parse()

	// Bütün flaglari ziyaret ederek, her hangi bir flag girilmediyse onu yakalayarak bilgisini ekrana veriyoruz
	flag.VisitAll(func(f *flag.Flag) {
		// f.Value.String() ile içerisinde gezdiğimiz bütün flagların içeriğine bakabiliyoruz.
		// == "" demek, boş olduğu anlamına gelir.
		if f.Value.String() == "" {
			fmt.Println(`"`, red(f.Name), `"`, "Parametresi girilmedi ! Lütfen bütün parametreleri doldurunuz !")
			os.Exit(1)
		}
	})

	if *urlFlag != "" || *apiFlag != "" {
		URLToID()
		getVideoInfo()
	}

}

var videoID string

// Verilen video URL'ni, API içinde kullanılabilmesi için, ID'ye dönüştüren fonksiyon.
// Örnek olarak video url: https://www.youtube.com/watch?v=t-fXDPE3lPE
// Örnek ID Çıktı: t-fXDPE3lPE
// ?v= parametresinden sonra gelen değeri ID olarak kabul ediyoruz
func URLToID() string {
	// url adındaki flagimizi yakalayarak içine verilen değeri string olarak alıyoruz ve videoUrl değişkenine atıyoruz
	videoURL := flag.Lookup("url").Value.(flag.Getter).Get().(string)
	// Regex'imizn Match olup olmadığını kontrol ediyoruz. match değişkeni boolean değer döndürecek.
	match, err := regexp.Match(`^.*((youtu.be\/)|(v\/)|(\/u\/\w\/)|(embed\/)|(watch\?))\??v?=?([^#&?]*).*`, []byte(videoURL))
	// Regeximizi compile ederek, elimize gelen url ile eşleştireceğiz
	r, _ := regexp.Compile(`^.*((youtu.be\/)|(v\/)|(\/u\/\w\/)|(embed\/)|(watch\?))\??v?=?([^#&?]*).*`)

	if match == false {
		fmt.Println(red("URL Match edilemedi !"))
	} else if err != nil {
		fmt.Println(red("Bir hata meydana geldi"))
	}

	// FindStringSubmatch fonksiyonunu kullanarak hem submatch ediyoruz url'i, hem de 8 item uzunluğuba sahip bir
	// listeye dönüştürüyoruz. Listemizin 7. elemenı bizim YouTube id'miz olmuş oluyor.
	urls := r.FindStringSubmatch(videoURL)

	videoID = urls[7]
	return videoID

}

// Video ile alakalı istatistekleri aldığımız fonksiyon
func getVideoInfo() string {
	var video Video

	var client http.Client

	// Flaglardan girilen değerleri yakalıyoruz.
	apiKey := flag.Lookup("api").Value.(flag.Getter).Get().(string)
	videoURL := videoID

	fullURL := "https://www.googleapis.com/youtube/v3/videos?key=" + apiKey + "&part=statistics&id=" + videoURL

	resp, err := client.Get(fullURL)

	if err != nil {
		panic(err)
	}

	// Eğer ki API'ye yaptığımız istek 200 status codunu dönerse yapacağımız işlemleri belirliyoruz
	if resp.StatusCode == http.StatusOK {
		// Gelen ilk response'de gelen değer byte'ler halinde geliyor. İstersek onu tip dönüşümü ile stringe çevirebiliriz.
		// Örnek: bodyString := string(bodyBytes)
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			panic(err)
		}
		// json değerini okuyoruz ve structaki değerlerle eşleştiriyoruz
		json.Unmarshal(bodyBytes, &video)

		for i := 0; i < len(video.Items); i++ {
			fmt.Println(green("Görüntülenme Sayısı: "), blue(video.Items[i].Statistics.ViewCount))
			fmt.Println(green("Beğeni Sayısı: "), blue(video.Items[i].Statistics.LikeCount))
			fmt.Println(green("Dislike Sayısı: "), blue(video.Items[i].Statistics.DislikeCount))
			fmt.Println(green("Yorum Sayısı:"), blue(video.Items[i].Statistics.CommentCount))
		}

	}

	return fullURL
}
