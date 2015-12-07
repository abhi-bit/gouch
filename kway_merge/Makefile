CC=gcc
ICU_CONFIG_EXECUTABLE=icu-config
ICU_INCLUDE_DIR=$(shell $(ICU_CONFIG_EXECUTABLE) --cppflags-searchpath)
ICU_LIB_SEARCHPATH=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-searchpath)
ICU_LIBS=$(shell $(ICU_CONFIG_EXECUTABLE) --ldflags-libsonly)
SOURCES=min_heap.c collate_json.c kway_merge.c kway_merge_test.c
OUTPUT=output
OUTPUT_GDB_DEBUG_FILES=$(OUTPUT).dSYM

all:
	$(CC) $(ICU_INCLUDE_DIR) $(ICU_LIB_SEARCHPATH) $(SOURCES) -o $(OUTPUT) $(ICU_LIBS)
	./$(OUTPUT)

clean:
	rm -rf $(OUTPUT) $(OUTPUT_GDB_DEBUG_FILES)

debug:
	$(CC) $(ICU_INCLUDE_DIR) $(ICU_LIB_SEARCHPATH) $(SOURCES) -o $(OUTPUT) -g $(ICU_LIBS)
