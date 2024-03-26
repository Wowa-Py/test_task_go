package main

import (
	"fmt"
	"sync"
	"time"
)

// Ttype представляет задачу с информацией о времени создания, выполнения и результате
type Ttype struct {
	id         int
	cT         time.Time // Используем time.Time для времени создания
	fT         time.Time // Используем time.Time для времени выполнения
	taskRESULT []byte
}

func main() {
	// Создаем канал для задач
	superChan := make(chan Ttype, 10)
	// Создаем каналы для выполненных и не выполненных задач
	doneTasks := make(chan Ttype, 10)
	undoneTasks := make(chan Ttype, 10)

	var wg sync.WaitGroup

	// Генератор задач
	taskCreator := func(a chan<- Ttype) {
		for {
			ct := time.Now()
			id := int(time.Now().UnixNano())
			// Условие появления ошибочных задач
			if ct.Nanosecond()%2 > 0 {
				a <- Ttype{id: id, cT: ct, taskRESULT: []byte("Some error occurred")}
			} else {
				a <- Ttype{id: id, cT: ct}
			}
			time.Sleep(10 * time.Millisecond) // Добавляем задержку для имитации работы
		}
	}

	go taskCreator(superChan)

	// Обработчик задач
	taskWorker := func(a Ttype) Ttype {
		if time.Since(a.cT) < 20*time.Second {
			a.taskRESULT = []byte("task has been succeeded")
		} else {
			a.taskRESULT = []byte("something went wrong")
		}
		a.fT = time.Now()
		time.Sleep(150 * time.Millisecond) // Имитация обработки задачи
		return a
	}

	// Сортировщик задач
	taskSorter := func(t Ttype) {
		if string(t.taskRESULT) == "task has been succeeded" {
			doneTasks <- t
		} else {
			undoneTasks <- t
		}
	}

	// Запускаем воркеры для обработки задач
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range superChan {
				result := taskWorker(t)
				taskSorter(result)
			}
		}()
	}

	// Закрываем каналы после завершения работы воркеров
	go func() {
		wg.Wait()
		close(doneTasks)
		close(undoneTasks)
	}()

	// Даем время на обработку задач
	time.Sleep(3 * time.Second)
	close(superChan)

	// Выводим результаты
	fmt.Println("Ошибки:")
	for t := range undoneTasks {
		fmt.Printf("Задача с ID %d, ошибка: %s\n", t.id, string(t.taskRESULT))
	}

	fmt.Println("Выполненные задачи:")
	for t := range doneTasks {
		fmt.Printf("Задача с ID %d успешно выполнена\n", t.id)
	}
}
