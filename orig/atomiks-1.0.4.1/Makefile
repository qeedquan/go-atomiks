CFLAGS = -std=gnu89 -O3 -Wall -Wextra -pedantic -Wno-long-long
LIBS = -lSDL2 -lSDL2_mixer

all: atomiks

atomiks: atomiks.o atomcore.o cfg.o drv_gra.o drv_inp.o drv_snd.o drv_tim.o gz.o
	gcc atomiks.o atomcore.o cfg.o drv_gra.o drv_inp.o drv_snd.o drv_tim.o gz.o $(LIBS) -o atomiks $(CFLAGS)

atomiks.o: atomiks.c
	gcc -c atomiks.c -o atomiks.o $(CFLAGS)

editor: editor.c atomcore.o drv_gra.o gz.o
	gcc editor.c atomcore.o drv_gra.o gz.o -lSDL2 -o editor $(CFLAGS)

file2c: file2c.c
	gcc $(CFLAGS) file2c.c -o file2c

png2bmp: png2bmp.c
	gcc $(CFLAGS) png2bmp.c -o png2bmp `sdl2-config --libs` -lSDL2_image

zopfli: zopfli-1.0/*.c
	gcc zopfli-1.0/*.c -O2 -W -Wall -Wextra -ansi -pedantic -lm -o zopfli

data.h: img/*.png snd/*.mod snd/*.wav zopfli file2c png2bmp
	echo "/* autogenerated file */" > data.h
	for x in img/*.png ; do ./png2bmp $$x ; done
	for x in img/*.bmp ; do ./zopfli $$x ; done
	rm img/*.bmp
	for x in img/*.bmp.gz ; do ./file2c $$x >> data.h ; done
	rm img/*.bmp.gz
	for x in snd/*.mod ; do ./file2c $$x >> data.h ; done
	for x in snd/*.wav ; do ./file2c $$x >> data.h ; done

levels.h: lev/lev*.dat
	echo "/* autogenerated file */" > levels.h
	for x in lev/*.dat ; do ./file2c $$x >> levels.h ; done

clean:
	rm -f editor atomiks atomiks.exe file2c png2bmp zopfli *.o
