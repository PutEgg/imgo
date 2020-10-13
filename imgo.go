package imgo

import (
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"fmt"
	"strconv"
)

const defaultCompareAccuracy = 10 //查找图片的精确值，默认查找图片平均有10分之一的像素对应即认为两部分图片一样。
const colorRange = 20000	//设置四个点颜色容差范围，解决视频压缩导致颜色误差问题

type Picture struct {
	Img             image.Image
	Width           int
	Height          int
	Path            string
	CompareAccuracy int
}

func NewJpeg(path string) (*Picture, error) {

	read, err := os.Open(path)
	if err != nil {
		return &Picture{}, err
	}
	defer read.Close()

	img, err := jpeg.Decode(read)
	if err != nil {
		return &Picture{}, err
	}

	return newPic(img, path), nil
}

func NewPng(path string) (*Picture, error) {

	read, err := os.Open(path)
	if err != nil {
		return &Picture{}, err
	}
	defer read.Close()

	img, err := png.Decode(read)
	if err != nil {
		return &Picture{}, err
	}

	return newPic(img, path), nil
}

func (p *Picture) SetCompareAccuracy(compareAccuracy int) {
	p.CompareAccuracy = compareAccuracy
}

func (p *Picture) SearchPic(searchPic *Picture) (bool, image.Rectangle) {
	rectangles := seekPos(p, searchPic, true)
	if len(rectangles) == 0 {
		return false, image.Rectangle{}
	}
	return true, rectangles[0]
}

func (p *Picture) SearchAllPic(searchPic *Picture) (bool, []image.Rectangle) {
	rectangles := seekPos(p, searchPic, false)
	if len(rectangles) == 0 {
		return false, rectangles
	}
	return true, rectangles
}

func (p *Picture) Replace(searchPic *Picture, replacer *Picture) (image.Image, error) {

	if searchPic.Width != replacer.Width || searchPic.Height != replacer.Height {
		return p.Img, errors.New("查找和替换的图片大小不一致")
	}

	isExist, rectangle := p.SearchPic(searchPic)
	if !isExist {
		return p.Img, errors.New("在" + p.Path + "并未发现" + searchPic.Path)
	}

	dst := p.Img
	if dst, ok := dst.(draw.Image); ok {
		draw.Draw(dst, rectangle, replacer.Img, image.Point{}, draw.Src)
	}
	return dst, nil
}

func (p *Picture) ReplaceAll(searchPic *Picture, replacer *Picture) (image.Image, error) {

	if searchPic.Width != replacer.Width || searchPic.Height != replacer.Height {
		return p.Img, errors.New("查找和替换的图片大小不一致")
	}

	isExist, rectangles := p.SearchAllPic(searchPic)
	if !isExist {
		return p.Img, errors.New("在" + p.Path + "并未发现" + searchPic.Path)
	}

	dst := p.Img
	if dst, ok := dst.(draw.Image); ok {
		for _, rectangle := range rectangles {
			draw.Draw(dst, rectangle, replacer.Img, image.Point{}, draw.Src)
		}
	}
	return dst, nil
}

func newPic(img image.Image, path string) *Picture {

	rectangle := img.Bounds()
	w := rectangle.Max.X
	h := rectangle.Max.Y
	return &Picture{
		Img:             img,
		Width:           w,
		Height:          h,
		Path:            path,
		CompareAccuracy: defaultCompareAccuracy,
	}
}

func scanAreaOk(intX, intY int, p, searchPic *Picture) bool {
	//fmt.Println("Found")
	h := searchPic.Height - 1
	w := searchPic.Width - 1

	if p.CompareAccuracy < 1 || h < p.CompareAccuracy || w < p.CompareAccuracy {
		p.SetCompareAccuracy(1)
	}
	
	
	
	for y := 0; y <= h; y += p.CompareAccuracy {
		for x := 0; x <= w; x += p.CompareAccuracy {
			
			var target_1_R,target_1_G,target_1_B = getRGB(p,intX+x,intY+y)
			var search_1_R,search_1_G,search_1_B = getRGB(searchPic,x,y)
		
			if (target_1_R-colorRange >= search_1_R || search_1_R >= target_1_R+colorRange) ||
				(target_1_G-colorRange >= search_1_G || search_1_G >= target_1_G+colorRange) ||
				(target_1_B-colorRange >= search_1_B || search_1_B >= target_1_B+colorRange) {
				return false
			}
		}
	}
	return true
}


func getRGB(pp *Picture,x int,y int)(int,int,int){
	var a,b,c,d = pp.Img.At(x, y).RGBA()
	var sa = fmt.Sprint(a)
	var sb = fmt.Sprint(b)
	var sc = fmt.Sprint(c)
	var sd = fmt.Sprint(d)
	_ = sd
	var ia,err1 = strconv.Atoi(sa)
	var ib,err2 = strconv.Atoi(sb)
	var ic,err3 = strconv.Atoi(sc)
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err1)
	}
	if err3 != nil {
		panic(err1)
	}
	return ia,ib,ic
}



func seekPos(p *Picture, searchPic *Picture, searchOnce bool) []image.Rectangle {
	var rectangles []image.Rectangle
	if searchPic.Width > p.Width || searchPic.Height > p.Height {
		return rectangles
	}
	
	var search_1_R,search_1_G,search_1_B = getRGB(searchPic,0,0)
	var search_2_R,search_2_G,search_2_B = getRGB(searchPic,searchPic.Width-1,searchPic.Height-1)
	var search_3_R,search_3_G,search_3_B = getRGB(searchPic,searchPic.Width-1,0)
	var search_4_R,search_4_G,search_4_B = getRGB(searchPic,0,searchPic.Height-1)
	
	for y := 0; y <= (p.Height - searchPic.Height); y++ {
		for x := 0; x <= (p.Width - searchPic.Width); x++ {
		
		
			//第一组
			
			var target_1_R,target_1_G,target_1_B = getRGB(p,x,y)
		
			//第二组
			
			var target_2_R,target_2_G,target_2_B = getRGB(p,x+searchPic.Width-1,y)
		
			//第三组
			
			var target_3_R,target_3_G,target_3_B = getRGB(p,x+searchPic.Width-1,y+searchPic.Height-1)
		
			//第四组
			
			var target_4_R,target_4_G,target_4_B = getRGB(p,x,y+searchPic.Height-1)
			


			
			if ((target_1_R-colorRange <= search_1_R && search_1_R <= target_1_R+colorRange) && (target_1_G-colorRange <= search_1_G && search_1_G <= target_1_G+colorRange) && (target_1_B-colorRange <= search_1_B && search_1_B <= target_1_B+colorRange)) &&
				((target_2_R-colorRange <= search_2_R && search_2_R <= target_2_R+colorRange) && (target_2_G-colorRange <= search_2_G && search_2_G <= target_2_G+colorRange) && (target_2_B-colorRange <= search_2_B && search_2_B <= target_2_B+colorRange)) &&
				((target_3_R-colorRange <= search_3_R && search_3_R <= target_3_R+colorRange) && (target_3_G-colorRange <= search_3_G && search_3_G <= target_3_G+colorRange) && (target_3_B-colorRange <= search_3_B && search_3_B <= target_3_B+colorRange)) &&
				((target_4_R-colorRange <= search_4_R && search_4_R <= target_4_R+colorRange) && (target_4_G-colorRange <= search_4_G && search_4_G <= target_4_G+colorRange) && (target_4_B-colorRange <= search_4_B && search_4_B <= target_4_B+colorRange)) { //四个角都在colorRange颜色范围内才继续下一步
				
				
			}else{
				
				continue
			}

			if !scanAreaOk(x, y, p, searchPic) { //四个角对上了在扫描区域，不成功直接下一次，
				
				continue
			}
			min := image.Point{X: x, Y: y}
			max := image.Point{X: x + searchPic.Width, Y: y + searchPic.Height}
			rectangles = append(rectangles, image.Rectangle{Min: min, Max: max})
			if searchOnce {
				return rectangles
			}
		}
	}
	return rectangles
}
