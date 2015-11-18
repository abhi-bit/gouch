CC=gcc
ICU_CONFIG_EXECUTABLE=icu-config
ICU_INCLUDE_DIR=$(shell $(ICU_CONFIG_EXECUTABLE) --cppflags-searchpath)
ICU_LIB_SEARCHPATH=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-searchpath)
ICU_LIBS=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-libsonly)
SOURCES=kway_merge_test.c min_heap.c collate_json.c

all:
	$(CC) $(ICU_INCLUDE_DIR) $(ICU_LIB_SEARCHPATH) $(SOURCES) -o test $(ICU_LIBS)
