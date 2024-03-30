cd proto
buf generate
cd ..

cp -r github.com/noble-assets/forwarding/* ./
rm -rf github.com
