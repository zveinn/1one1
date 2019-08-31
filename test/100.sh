echo "Spawning 100 processes"
cd "../collector"
for i in {1..100}
do
    ( go run main.go $i & )
done