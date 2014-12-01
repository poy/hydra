package hydra_test

import (
	"encoding"
	. "github.com/apoydence/hydra"
	. "github.com/apoydence/hydra/testing_helpers"
	. "github.com/apoydence/hydra/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scaffolding", func() {
	Context("Integrate", func() {
		It("with a linear path", func(done Done) {
			defer close(done)
			doneChan := make(chan interface{})
			results := make(chan encoding.BinaryMarshaler)
			wrapperConsumer := func(s SetupFunction) {
				consumer(s, 7, results, doneChan)
			}

			go func() {
				defer close(results)
				for i := 0; i < 7; i++ {
					<-doneChan
				}
			}()
			go NewSetupScaffolding()(producer, filter, wrapperConsumer)

			expectedData := [...]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
			rxData := make([]int, 0)
			for data := range results {
				rxData = append(rxData, (data.(IntMarshaler)).Number)
			}
			Expect(rxData).To(ConsistOf(expectedData))
		}, 1)

		It("with a non-linear path", func(done Done) {
			defer close(done)

			doneChan1 := make(chan interface{})
			results1 := make(chan encoding.BinaryMarshaler)
			wrapperConsumer1 := func(s SetupFunction) {
				consumer(s, 7, results1, doneChan1)
			}

			go func() {
				defer close(results1)
				for i := 0; i < 7; i++ {
					<-doneChan1
				}
			}()

			doneChan2 := make(chan interface{})
			results2 := make(chan encoding.BinaryMarshaler)
			wrapperConsumer2 := func(s SetupFunction) {
				consumer2(s, 17, results2, doneChan2)
			}

			go func() {
				defer close(results2)
				for i := 0; i < 17; i++ {
					<-doneChan2
				}
			}()

			go NewSetupScaffolding()(producer, filter, filter2, wrapperConsumer1, wrapperConsumer2)

			go func() {
				expectedIndex := 0
				expectedData := [...]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
				for data := range results1 {
					Expect(expectedData[expectedIndex]).To(Equal(data.(IntMarshaler).Number))
					expectedIndex++
				}
			}()

			expectedData := [...]int{0, 2, 4, 6, 8}
			rxData := make([]int, 0)
			for data := range results2 {
				rxData = append(rxData, data.(IntMarshaler).Number)
			}
			Expect(rxData).To(ConsistOf(expectedData))
		}, 1)
	})
})

func producer(s SetupFunction) {
	out := s.AsProducer().Build()
	defer close(out)
	for i := 0; i < 10; i++ {
		out <- NewIntMarshaler(i)
	}
}

func filter(s SetupFunction) {
	in, out := s.Instances(10).AsFilter("github.com/apoydence/hydra_test.producer").Build()
	defer close(out)

	for data := range in {
		out <- data
	}
}

func filter2(s SetupFunction) {
	in, out := s.Instances(5).AsFilter("github.com/apoydence/hydra_test.producer").Build()
	defer close(out)

	for data := range in {
		if data.(IntMarshaler).Number%2 == 0 {
			out <- data
		}
	}
}

func consumer(s SetupFunction, count int, results WriteOnlyChannel, doneChan chan interface{}) {
	defer func() {
		doneChan <- nil
	}()
	in := s.Instances(count).AsConsumer("github.com/apoydence/hydra_test.filter").Build()
	for data := range in {
		results <- data
	}
}

func consumer2(s SetupFunction, count int, results WriteOnlyChannel, doneChan chan interface{}) {
	defer func() {
		doneChan <- nil
	}()
	in := s.Instances(count).AsConsumer("github.com/apoydence/hydra_test.filter2").Build()
	for data := range in {
		results <- data
	}
}
