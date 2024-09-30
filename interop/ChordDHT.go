//go:build interop

package Interop

/*
#cgo CFLAGS: -I/usr/lib/jvm/java-11-openjdk-amd64/include -I/usr/lib/jvm/java-11-openjdk-amd64/include/linux
#cgo LDFLAGS: -L/usr/lib/jvm/java-11-openjdk-amd64/lib/amd64/server -ljvm
#include <jni.h> // The Java Native Interface header
#include <stdlib.h>
#include <string.h>

// Declare a global variable to hold the JVM instance
JavaVM* jvm;

char* init_jvm()
{
	JavaVMInitArgs vm_args; // Initialization arguments
	vm_args.version = JNI_VERSION_1_6; // set the JNI version
	vm_args.nOptions = 0; // no options (like class path)
	vm_args.ignoreUnrecognized = 0;
	JNIEnv* env;
	int res = JNI_CreateJavaVM(&jvm, (void**)&env, &vm_args);
	if (res < 0) {
		char* error_msg;
		switch(res) {
			case JNI_ERR:
				error_msg = "unknown error";
				break;
			case JNI_EDETACHED:
				error_msg = "thread detached from the VM";
				break;
			case JNI_EVERSION:
				error_msg = "JNI version error";
				break;
			case JNI_ENOMEM:
				error_msg = "not enough memory";
				break;
			case JNI_EEXIST:
				error_msg = "VM already created";
				break;
			case JNI_EINVAL:
				error_msg = "invalid arguments";
				break;
		default:
			error_msg = "unknown error code";
		}
	return error_msg;
	}
	return NULL;
}


// Global variables
jclass chordClass;
jmethodID chordConstructorNewChord;
jmethodID chordConstructorJoinChord;
jmethodID methodSet;
jmethodID methodGet;
jmethodID methodDelete;
jmethodID methodGetAllKeys;


char* get_env(JNIEnv** env)
{
 	// Get the JVM environment for the current thread
	jint res = (*jvm)->GetEnv(jvm, (void**)env, JNI_VERSION_1_6);

	if(res == JNI_OK) { // All is well, JVM is attached
		return NULL;
	}
	// Thread is not attached to the JVM
	else if (res == JNI_EDETACHED)
	{
		// Attach the thread to the JVM
		res = (*jvm)->AttachCurrentThread(jvm, (void **)env, NULL);
		if (res != JNI_OK) { // Failed to attach thread to JVM
			return strdup("Failed to attach thread to JVM");
		}
		return NULL; // All is well
	}
	else
	{
		// Some other error
		return strdup("Failed to get JNIEnv");
	}
}


// Function to get the exception message
char* get_exception_message(JNIEnv* env)
{
	// get the exception from the JVM
	jthrowable exception = (*env)->ExceptionOccurred(env);
	(*env)->ExceptionClear(env); // Clear the exception

	// All exceptions in Java are Throwable,
	// so we can cast "exception" it to Throwable
	jclass classThrowable = (*env)->FindClass(env, "java/lang/Throwable");

	// get "String toString()" method from Throwable
	jmethodID methodToString = (*env)->GetMethodID(env, classThrowable, "toString", "()Ljava/lang/String;");

	// call toString() method on the exception
	jstring message = (jstring)(*env)->CallObjectMethod(env, exception, methodToString);

	// convert to C string
	const char* messageChars = (*env)->GetStringUTFChars(env, message, NULL);

	// make a copy of the message
	char* messageCopy = strdup(messageChars);

	// clean up
	(*env)->ReleaseStringUTFChars(env, message, messageChars);
	(*env)->DeleteLocalRef(env, message);
	(*env)->DeleteLocalRef(env, classThrowable);

	(*env)->DeleteLocalRef(env, exception);
	return messageCopy;
}


char* load_chord_class()
{
	JNIEnv* env = NULL;
	char* error = get_env(&env);
	if (error != NULL) {
		return error;
 	}

	chordClass = NULL;
	chordConstructorNewChord = NULL;
	chordConstructorJoinChord = NULL;
	methodSet = NULL;
	methodGet = NULL;
	methodDelete = NULL;
	methodGetAllKeys = NULL;

	jclass urlClass = (*env)->FindClass(env, "java/net/URL");
	if (urlClass == NULL)
	{
		error = "Could not find URL class";
		goto cleanup;
	}

	// load the constructor
	jmethodID urlConstructor = (*env)->GetMethodID(env, urlClass, "<init>", "(Ljava/lang/String;)V");
	if (urlConstructor == NULL)
	{
		error = "Could not find URL constructor";
		goto cleanup;
	}

	// Create the value of the first parameter to the constructor
 	jstring urlStr = (*env)->NewStringUTF(env, "file:///workspaces/AAG/interop/");

	// Call the constructor and create an instance of URL
	jobject urlObj = (*env)->NewObject(env, urlClass, urlConstructor, urlStr);
	if (urlObj == NULL)
	{
		error = "Could not create instance of URL object";
		goto cleanup;
 	}

	jobjectArray urlArray = (*env)->NewObjectArray(env, 1, urlClass, urlObj);
	if (urlObj == NULL)
	{
		error = "Could not create instance of URL[]{urlObj}";
		goto cleanup;
	}

	jclass urlClassLoaderClass = (*env)->FindClass(env, "java/net/URLClassLoader");
	if (urlClassLoaderClass == NULL)
	{
		error = "Could not find URLClassLoader class";
		goto cleanup;
	}

	jmethodID urlClassLoaderConstructor = (*env)->GetMethodID(env, urlClassLoaderClass, "<init>", "([Ljava/net/URL;)V");
	if (urlClassLoaderConstructor == NULL)
	{
		error = "Could not find URLClassLoader constructor";
		goto cleanup;
 	}

	jobject urlClassLoader = (*env)->NewObject(env, urlClassLoaderClass, urlClassLoaderConstructor, urlArray);
	jmethodID loadClassMethod = (*env)->GetMethodID(env, urlClassLoaderClass, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;");
	if (loadClassMethod == NULL)
	{
		error = "Could not find loadClass method";
		goto cleanup;
	}

	jstring className = (*env)->NewStringUTF(env, "dht.Chord");
	jclass localChordClass = (jclass)(*env)->CallObjectMethod(env, urlClassLoader, loadClassMethod, className);
	if (localChordClass == NULL)
	{
		if((*env)->ExceptionCheck(env))
		{
			error = get_exception_message(env);
		}
		else
		{
			error = "Could not load Chord class";
		}
		goto cleanup;
	}

	// set the global reference
 	chordClass = (*env)->NewGlobalRef(env, localChordClass);


	// Create new Chord
	chordConstructorNewChord = (*env)->GetMethodID(env, chordClass, "<init>", "(Ljava/lang/String;I)V");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}

	// Join Chord
	chordConstructorJoinChord = (*env)->GetMethodID(env, chordClass, "<init>", "(Ljava/lang/String;Ljava/lang/String;I)V");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}

	// Set
	methodSet = (*env)->GetMethodID(env, chordClass, "set", "(Ljava/lang/String;Ljava/lang/String;)V");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}

	// Get
	methodGet = (*env)->GetMethodID(env, chordClass, "get", "(Ljava/lang/String;)Ljava/lang/String;");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}

	// Delete
	methodDelete = (*env)->GetMethodID(env, chordClass, "delete", "(Ljava/lang/String;)V");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}

	// Get All Keys
	methodGetAllKeys = (*env)->GetMethodID(env, chordClass, "getAllKeys", "()[Ljava/lang/String;");
	if ((*env)->ExceptionCheck(env)) {
		error = get_exception_message(env);
		goto cleanup;
	}
cleanup:
	if (urlClass != NULL) (*env)->DeleteLocalRef(env, urlClass);
	if (urlObj != NULL) (*env)->DeleteLocalRef(env, urlObj);
	if (urlArray != NULL) (*env)->DeleteLocalRef(env, urlArray);
	if (urlClassLoaderClass != NULL) (*env)->DeleteLocalRef(env, urlClassLoaderClass);
	if (urlClassLoader != NULL) (*env)->DeleteLocalRef(env, urlClassLoader);
	if (className != NULL) (*env)->DeleteLocalRef(env, className);
	if (localChordClass != NULL) (*env)->DeleteLocalRef(env, localChordClass);

	return error;
}


jobject call_chord_constructor_new_chord(char* node_name, int port, char** out_error)
{
	jobject chordObject = NULL;
	jstring jnodeName = NULL;

	// get environment
	JNIEnv* env;
	char* error = get_env(&env);
	if (error != NULL)
	{
		*out_error = error;
		return NULL;
	}

	// Convert the C string to a Java string
	jnodeName = (*env)->NewStringUTF(env, node_name);
	if ((*env)->ExceptionCheck(env))
	{
		*out_error = get_exception_message(env);
		goto cleanup;
	}

	// Call the constructor using NewObject function
	chordObject = (*env)->NewObject(env, chordClass, chordConstructorNewChord, jnodeName, port);
	if ((*env)->ExceptionCheck(env))
	{
		*out_error = get_exception_message(env);
		goto cleanup;
	}

cleanup:
	// Free local references
	if (jnodeName) {(*env)->DeleteLocalRef(env, jnodeName);}
		return chordObject;
}


jobject call_chord_constructor_join_chord(char* node_name, char* root_name, int port, char** out_error)
{
	jobject newChordObject = NULL;
	jstring jnodeName = NULL;
	jstring jrootName = NULL;

	// get environment
	JNIEnv* env = NULL;
	char* error = get_env(&env);
	if (error != NULL)
	{
		*out_error = error;
		return NULL;
	}

	// Convert the C strings to Java strings
	jnodeName = (*env)->NewStringUTF(env, node_name);
	jrootName = (*env)->NewStringUTF(env, root_name);
	if ((*env)->ExceptionCheck(env))
	{
		*out_error = get_exception_message(env);
		goto cleanup;
	}
		// Call the constructor
	newChordObject = (*env)->NewObject(env, chordClass, chordConstructorJoinChord, jnodeName, jrootName, port);
	if ((*env)->ExceptionCheck(env))
	{
		*out_error = get_exception_message(env);
		goto cleanup;
	}

cleanup:
	// Free local references
	if (jnodeName) (*env)->DeleteLocalRef(env, jnodeName);
	if (jrootName) (*env)->DeleteLocalRef(env, jrootName);

	return newChordObject;
}


void call_method_set(jobject chordObject, char* key, char* value, char** out_error)
{
	jstring jkey = NULL;

	jstring jvalue = NULL;
	JNIEnv* env;
	char* error = get_env(&env);
	if (error != NULL) {
		*out_error = error;
		return;
	}

	// Convert the C strings to Java strings
	jkey = (*env)->NewStringUTF(env, key);
	jvalue = (*env)->NewStringUTF(env, value);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
		goto cleanup;
	}

	// Call the set method - As the method returns void, we use CallVoidMethod
	(*env)->CallVoidMethod(env, chordObject, methodSet, jkey, jvalue);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
	}

cleanup:
	// Free local references
	if (jkey) (*env)->DeleteLocalRef(env, jkey);
	if (jvalue) (*env)->DeleteLocalRef(env, jvalue);
}


char* call_method_get(jobject chordObject, char* key, char** out_error)
{
	jstring jkey = NULL;
	jstring jresult = NULL;
	const char* result = NULL;

	JNIEnv* env;
	char* error = get_env(&env);
	if (error != NULL) {
		return error;
	}

	// Convert the C string to a Java string
	jkey = (*env)->NewStringUTF(env, key);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
		goto cleanup;
	}

	// Call the get method
	jresult = (jstring)(*env)->CallObjectMethod(env, chordObject, methodGet, jkey);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
		goto cleanup;
	}

	// not found
	if(jresult == NULL) {
		goto cleanup;
	}

	// Convert the Java string to a C string
	result = (*env)->GetStringUTFChars(env, jresult, 0);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
	}

cleanup:
	// Free local references
	if (jkey) (*env)->DeleteLocalRef(env, jkey);
	if (jresult) (*env)->DeleteLocalRef(env, jresult);

	return (char*)result;
}


void call_method_delete(jobject chordObject, char* key, char** out_error){

	JNIEnv* env;
    char* error = get_env(&env);
    if (error != NULL) {
        *out_error = error;
        return;
    }

    jstring jkey = (*env)->NewStringUTF(env, key);
    if (jkey == NULL) {
        *out_error = strdup("Failed to create jstring for key");
        return;
    }

    (*env)->CallVoidMethod(env, chordObject, methodDelete, jkey);
    if ((*env)->ExceptionCheck(env)) {
        *out_error = get_exception_message(env);
        (*env)->DeleteLocalRef(env, jkey);
        return;
    }

    (*env)->DeleteLocalRef(env, jkey);
}


char** call_method_get_all_keys(jobject chordObject, char** out_error)
{
	jobjectArray jresult = NULL;
	char** result = NULL;

	JNIEnv* env;
	char* error = get_env(&env);
	if (error != NULL)
	{
		*out_error = error;
		return NULL;
	}

	jsize len;
	int i;

	// Call the getAllKeys method
	jresult = (jobjectArray)(*env)->CallObjectMethod(env, chordObject, methodGetAllKeys);
	if ((*env)->ExceptionCheck(env))
	{
		*out_error = get_exception_message(env);
		goto cleanup;
	}

	// Convert the Java string array to a C string array
	len = (*env)->GetArrayLength(env, jresult);

	////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// TODO: copy the jresult to “result”.
	// Make sure you allocate enough space!
	// Make sure the LAST element in the array is NULL – to mark it is the end of the array

	result = (char**)malloc((len + 1) * sizeof(char*)); // Allocate enough space for all strings + NULL terminator
	if (result == NULL) {
        *out_error = strdup("Failed to allocate result");
        goto cleanup;
	}

	for (int i = 0; i <= len; i++) {
		result[i] = NULL;
	}

    for (int i = 0; i < len; i++) {
        jstring jstr = (jstring)(*env)->GetObjectArrayElement(env, jresult, i);
        if (jstr == NULL) {
            *out_error = strdup("Failed to get jstring from jobjectArray");
            goto cleanup;
        }

        const char* cstr = (*env)->GetStringUTFChars(env, jstr, NULL);
        if (cstr == NULL) {
            *out_error = strdup("Failed to convert jstring to C string");
            (*env)->DeleteLocalRef(env, jstr);
            goto cleanup;
        }

        result[i] = strdup(cstr); // Duplicate the C string
        if (result[i] == NULL) {
            *out_error = strdup("Failed to duplicate C string");
            (*env)->ReleaseStringUTFChars(env, jstr, cstr);
            (*env)->DeleteLocalRef(env, jstr);
            goto cleanup;
        }

        (*env)->ReleaseStringUTFChars(env, jstr, cstr);
        (*env)->DeleteLocalRef(env, jstr);
    }


cleanup:
	// Free local references
	if (jresult) (*env)->DeleteLocalRef(env, jresult);

	// Clean up on error
    if (*out_error != NULL && result != NULL) {
        for (int i = 0; i <= len; i++) {
            if (result[i] != NULL) {
                free(result[i]);
            }
        }
		free(result);
        return NULL;
    }

	return result;
}


jboolean get_is_first_field(jobject chordObject, char** out_error)
{
	jfieldID fieldIsFirst;
	jboolean isFirst;

	JNIEnv* env;
	char* error = get_env(&env);

	if (error != NULL) {
		*out_error = error;
		return JNI_FALSE;
	}

	// Get the field ID
	// "Z" is the JNI signature for boolean
	fieldIsFirst = (*env)->GetFieldID(env, chordClass, "isFirst", "Z");
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
		return JNI_FALSE;
	}

	// Get the field value
	isFirst = (*env)->GetBooleanField(env, chordObject, fieldIsFirst);
	if ((*env)->ExceptionCheck(env)) {
		*out_error = get_exception_message(env);
		return JNI_FALSE;
	}

	return isFirst;
}


char* delete_global_ref(void* obj)
{
	JNIEnv* env;
	char* error = get_env(&env);
	if (error != NULL) {
		return strdup(obj);
	}

	(*env)->DeleteGlobalRef(env, (jobject)obj);
	return NULL;
}

char* get_string_from_array(char** array, int index){
	 if (array == NULL || array[index] == NULL) {
        return NULL;
    }
    return array[index];
}


*/
import "C"
import (
	"errors"
	"unsafe"
)

var is_jvm_initialized bool = false

func LoadJVM() error {
	// This is not thread-safe, but it's fine for our example
	if !is_jvm_initialized {
		err := C.init_jvm() // Load JVM
		if err != nil {
			return errors.New(C.GoString(err))
		}
		is_jvm_initialized = true

		err = C.load_chord_class() // Load Chord and its Methods
		if err != nil {
			return errors.New(C.GoString(err))
		}
	}
	return nil
}

type ChordDHT struct {
	// holds the jobject returned from the constructor
	instance unsafe.Pointer // Go unsafe.Pointer == C void*
}

func NewChordDHT(rootName string, port int) (*ChordDHT, error) {
	var out_error *C.char
	rootNameC := C.CString(rootName)        // Convert Go string to C string
	defer C.free(unsafe.Pointer(rootNameC)) // Free the C string when the function returns

	// Call the C function
	chordObject := C.call_chord_constructor_new_chord(rootNameC, C.int(port), &out_error)
	if out_error != nil { // Check if there's an error
		return nil, errors.New(C.GoString(out_error))
	}

	// Return the ChordDHT object
	return &ChordDHT{instance: unsafe.Pointer(chordObject)}, nil
}

// /////////////////////////////
func JoinChordDHT(nodeName string, rootName string, port int) (*ChordDHT, error) {
	var out_error *C.char

	// Convert Go strings to C strings
	nodeNameC := C.CString(nodeName)
	rootNameC := C.CString(rootName)
	defer C.free(unsafe.Pointer(nodeNameC)) // Free the C string when the function returns
	defer C.free(unsafe.Pointer(rootNameC)) // Free the C string when the function returns

	// Call the C function
	chordObject := C.call_chord_constructor_join_chord(nodeNameC, rootNameC, C.int(port), &out_error)
	if out_error != nil { // Check if there's an error
		return nil, errors.New(C.GoString(out_error))
	}

	// Return the ChordDHT object
	return &ChordDHT{instance: unsafe.Pointer(chordObject)}, nil
}

// receiver "chord" is a pointer to a ChordDHT object
func (chord *ChordDHT) GetIsFirst() (bool, error) {
	var out_error *C.char

	// call C function
	isFirst := C.get_is_first_field(C.jobject(chord.instance), &out_error)
	if out_error != nil {
		return false, errors.New(C.GoString(out_error))
	}

	// convert C boolean to Go boolean
	if isFirst != C.JNI_FALSE {
		return true, nil
	} else {
		return false, nil
	}
}

func (chord *ChordDHT) Get(key string) (string, error) {
	var out_error *C.char
	keyC := C.CString(key)
	defer C.free(unsafe.Pointer(keyC))

	valueC := C.call_method_get(C.jobject(chord.instance), keyC, &out_error)
	if out_error != nil {
		return "", errors.New(C.GoString(out_error))
	}

	return C.GoString(valueC), nil
}

func (chord *ChordDHT) GetAllKeys() ([]string, error) {
	var out_error *C.char

	keysC := C.call_method_get_all_keys(C.jobject(chord.instance), &out_error)
	if out_error != nil {
		return nil, errors.New(C.GoString(out_error))
	}

	// Convert the C array of strings to a Go slice of strings
	keys := make([]string, 0)
	var i int
	for {
		keyC := C.get_string_from_array(keysC, C.int(i))
		if keyC == nil {
			break
		}
		keys = append(keys, C.GoString(keyC))
		C.free(unsafe.Pointer(keyC))
		i++
	}
	C.free(unsafe.Pointer(keysC))

	return keys, nil
}

// /////////////////////////////
func (chord *ChordDHT) Set(key string, value string) error {
	var out_error *C.char
	keyC := C.CString(key)
	valueC := C.CString(value)
	defer C.free(unsafe.Pointer(keyC))
	defer C.free(unsafe.Pointer(valueC))

	C.call_method_set(C.jobject(chord.instance), keyC, valueC, &out_error)
	if out_error != nil {
		return errors.New(C.GoString(out_error))
	}
	return nil
}

func (chord *ChordDHT) Delete(key string) error {
	var out_error *C.char
	keyC := C.CString(key)
	defer C.free(unsafe.Pointer(keyC))

	C.call_method_delete(C.jobject(chord.instance), keyC, &out_error)
	if out_error != nil {
		return errors.New(C.GoString(out_error))
	}
	return nil
}

func (dht *ChordDHT) DeleteObject() error {
	// Call the C function to delete the global reference
	out_error := C.delete_global_ref(dht.instance)
	if out_error != nil {
		return errors.New(C.GoString(out_error))
	}
	return nil
}
