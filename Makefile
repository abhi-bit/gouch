CC=gcc
ICU_CONFIG_EXECUTABLE=icu-config
ICU_INCLUDE_DIR=$(shell $(ICU_CONFIG_EXECUTABLE) --cppflags-searchpath)
ICU_LIB_SEARCHPATH=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-searchpath)
ICU_LIBS=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-libsonly)
COLLATE_JSON_OUT=kway_merge/collate_json.o
COLLATE_JSON_SRC=kway_merge/collate_json.c
COLLATE_JSON_SHARED_LIB=libcollate_json.so

all:
	$(CC) $(ICU_INCLUDE_DIR) -g -fPIC -c -o $(COLLATE_JSON_OUT) $(COLLATE_JSON_SRC)
	$(CC) $(ICU_INCLUDE_DIR) $(ICU_LIB_SEARCHPATH) -g -fPIC -shared -o \
		$(COLLATE_JSON_SHARED_LIB) $(ICU_LIBS)
	cp $(COLLATE_JSON_SHARED_LIB) http_server/

