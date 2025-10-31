package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

const TotalPoints = 1000000

// sequentialPi обчислює PI послідовно в одному потоці.
func sequentialPi(numPoints int) float64 {
	rand.Seed(time.Now().UnixNano())
	insideCircle := 0

	for i := 0; i < numPoints; i++ {
		x := rand.Float64()
		y := rand.Float64()

		// Перевіряємо, чи точка потрапила в коло з радіусом 1
		if x*x+y*y <= 1.0 {
			insideCircle++
		}
	}

	// PI ≈ 4 * (Кількість точок в колі / Загальна кількість точок)
	return 4.0 * float64(insideCircle) / float64(numPoints)
}

// worker обчислює PI для заданої кількості точок і надсилає результат в канал.
// Використовує окремий генератор rand для кожної горутини, щоб уникнути race condition.
func worker(numPoints int, resultChan chan int) {
	// Для кожної горутини використовується окремий rand.Source,
	// щоб уникнути синхронізації при генерації випадкових чисел.
	// Час тут використовується як простий спосіб отримати унікальне зерно.
	source := rand.NewSource(time.Now().UnixNano() + int64(numPoints))
	r := rand.New(source)

	insideCircle := 0
	for i := 0; i < numPoints; i++ {
		x := r.Float64()
		y := r.Float64()

		if x*x+y*y <= 1.0 {
			insideCircle++
		}
	}

	// Відправка результату (кількість точок в колі) в канал
	resultChan <- insideCircle
}

// parallelPi обчислює PI, розбиваючи роботу на numThreads горутин.
func parallelPi(totalPoints, numThreads int) (float64, time.Duration) {
	startTime := time.Now()

	// Встановлення максимальної кількості використовуваних ядер
	runtime.GOMAXPROCS(numThreads)

	// Розподіл точок між потоками
	pointsPerWorker := totalPoints / numThreads
	remainder := totalPoints % numThreads

	resultChan := make(chan int, numThreads) // Канал для збору результатів
	var wg sync.WaitGroup                    // WaitGroup для з'єднання горутин

	// Запуск горутин
	for i := 0; i < numThreads; i++ {
		currentPoints := pointsPerWorker
		if i < remainder {
			currentPoints++ // Додаємо залишок першим потокам
		}

		wg.Add(1)
		go func(pts int) {
			defer wg.Done()
			worker(pts, resultChan)
		}(currentPoints)
	}

	// Асинхронне закриття каналу після завершення всіх горутин (аналог join)
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Збір результатів з каналу
	totalInside := 0
	for count := range resultChan {
		totalInside += count
	}

	elapsedTime := time.Since(startTime)

	// Фінальне обчислення PI
	pi := 4.0 * float64(totalInside) / float64(totalPoints)
	return pi, elapsedTime
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // Використовувати всі доступні ядра за замовчуванням
	fmt.Println("Обчислення числа PI методом Монте-Карло")
	fmt.Printf("Загальна кількість точок: %d\n\n", TotalPoints)

	fmt.Println("--- Послідовне обчислення (один потік) ---")
	startTimeSeq := time.Now()
	piSeq := sequentialPi(TotalPoints)
	elapsedTimeSeq := time.Since(startTimeSeq)
	fmt.Printf("Отримане PI: %.6f\n", piSeq)
	fmt.Printf("Час обчислення: %s\n", elapsedTimeSeq)

	fmt.Println("\n--- Паралельне обчислення (різна кількість потоків) ---")
	report := "**Звіт про залежність часу обчислення від кількості потоків:**\n\n"
	report += "| Кількість Потоків | Отримане PI | Час Обчислення (мс) |\n"

	report += fmt.Sprintf("| 1 (Послідовно)   | %.6f   | %.2f |\n", piSeq, float64(elapsedTimeSeq.Microseconds())/1000.0)

	threadCounts := []int{2, 4, 8, 16, 32, 64}

	for _, numThreads := range threadCounts {
		piPar, elapsedTimePar := parallelPi(TotalPoints, numThreads)

		fmt.Printf("Кількість потоків: %d\n", numThreads)
		fmt.Printf("Отримане PI: %.6f\n", piPar)
		fmt.Printf("Час обчислення: %s\n", elapsedTimePar)

		report += fmt.Sprintf("| %d                  | %.6f   | %.2f |\n", numThreads, piPar, float64(elapsedTimePar.Microseconds())/1000.0)
	}

	fmt.Println("\n--- Загальний результат ---")
	fmt.Println(report)
}
