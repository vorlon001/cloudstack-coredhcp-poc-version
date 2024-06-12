git tag:b510961d3c256d79ae1658f0d5d3dd065348dd5e
GitRepo: github.com/spf13/viper
diff --git a/viper.go b/viper.go
index 4bbc7db..80b1886 100644
--- a/viper.go
+++ b/viper.go
@@ -1541,7 +1541,21 @@ func (v *Viper) Set(key string, value any) {
 // and key/value stores, searching in one of the defined paths.
 func ReadInConfig() error { return v.ReadInConfig() }
 
-func (v *Viper) ReadInConfig(buffer []byte) error {
+func (v *Viper) ReadInConfigfromBuffer(format string, buffer []byte) error {
+
+        config := make(map[string]any)
+
+        err := v.unmarshalReaderBuffer(format, buffer, config)
+        if err != nil {
+                return err
+        }
+
+        v.config = config
+        return nil
+}
+
+
+func (v *Viper) ReadInConfig() error {
 	
 	v.logger.Info("attempting to read in config file")
 	filename, err := v.getConfigFile()
@@ -1561,7 +1575,6 @@ func (v *Viper) ReadInConfig(buffer []byte) error {
 	
 	config := make(map[string]any)
 
-	//err = v.unmarshalReader(bytes.NewReader(file), config)
 	err = v.unmarshalReader(bytes.NewReader(file), config)
 	if err != nil {
 		return err
@@ -1711,6 +1724,18 @@ func (v *Viper) writeConfig(filename string, force bool) error {
 	return f.Sync()
 }
 
+func (v *Viper) unmarshalReaderBuffer(format string, buffer []byte, c map[string]any) error {
+
+	err := v.decoderRegistry.Decode(format, buffer, c)
+	if err != nil {
+		return ConfigParseError{err}
+	}
+
+        insensitiviseMap(c)
+        return nil
+}
+
+
 func (v *Viper) unmarshalReader(in io.Reader, c map[string]any) error {
 	buf := new(bytes.Buffer)
 	buf.ReadFrom(in)
