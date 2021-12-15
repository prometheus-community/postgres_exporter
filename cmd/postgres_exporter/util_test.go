package postgres_exporter_test

import (
	"fmt"
	"math"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/thiagosantosleite/postgres_exporter/cmd/postgres_exporter"
)

var _ = Describe("Util", func() {
	Context("ParseFingerprint", func() {
		var (
			ctrl *gomock.Controller
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("Should pass if ParseFingerprint parse correctly", func() {
			ret, _ := ParseFingerprint("postgresql://user:password@host:5432/posgres")
			Expect(ret).To(Equal("host:5432"))
		})

		It("Should pass if ParseFingerprint works", func() {
			_, err := ParseFingerprint("postgresql://user:password@host:5432/posgres")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should fail if invalid url", func() {
			_, err := ParseFingerprint("postgresql://xxxxxx")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("DbToString", func() {
		It("Should pass if DbToString parse int64", func() {
			ret, ok := DbToString(int64(1))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal("1"))
		})

		It("Should pass if DbToString parse float64", func() {
			ret, ok := DbToString(float64(1.0))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal("1"))
		})

		It("Should pass if DbToString parse time.Time", func() {
			ret, ok := DbToString(time.Now())
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(fmt.Sprintf("%d", time.Now().Unix())))
		})

		It("Should pass if DbToString parse nil", func() {
			ret, ok := DbToString(nil)
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(""))
		})

		It("Should pass if DbToString parse []byte", func() {
			ret, ok := DbToString([]byte{65, 66})
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal("AB"))
		})

		It("Should pass if DbToString parse string", func() {
			ret, ok := DbToString("abc")
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal("abc"))
		})
		It("Should pass if DbToString parse bool", func() {
			ret, ok := DbToString(true)
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal("true"))
		})
		It("Should fail if DbToString fails to parse", func() {
			ret, ok := DbToString(fmt.Errorf("dummy"))
			Expect(ok).To(BeFalse())
			Expect(ret).To(Equal(""))
		})
	})

	Context("DbToUint64", func() {
		It("Should pass if DbToUint64 parse int64", func() {
			ret, ok := DbToUint64(int64(1))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(1)))
		})

		It("Should pass if DbToUint64 parse float64", func() {
			ret, ok := DbToUint64(float64(1.0))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(1)))
		})

		It("Should pass if DbToUint64 parse time.Time", func() {
			ret, ok := DbToUint64(time.Now())
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(time.Now().Unix())))
		})

		It("Should pass if DbToUint64 parse nil", func() {
			ret, ok := DbToUint64(nil)
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(0)))
		})

		It("Should pass if DbToUint64 parse []byte", func() {
			ret, ok := DbToUint64([]byte{33})
			Expect(ok).To(BeFalse())
			Expect(ret).To(Equal(uint64(0)))
		})

		It("Should pass if DbToUint64 parse string", func() {
			ret, ok := DbToUint64("123")
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(123)))
		})
		It("Should pass if DbToUint64 parse bool", func() {
			ret, ok := DbToUint64(true)
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(uint64(1)))
		})
		It("Should fail if DbToUint64 fails to parse", func() {
			ret, ok := DbToUint64(fmt.Errorf("dummy"))
			Expect(ok).To(BeFalse())
			Expect(ret).To(Equal(uint64(0)))
		})
	})

	Context("DbToFloat64", func() {
		It("Should pass if DbToFloat64 parse int64", func() {
			ret, ok := DbToFloat64(int64(1))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(float64(1)))
		})

		It("Should pass if DbToFloat64 parse float64", func() {
			ret, ok := DbToFloat64(float64(1.0))
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(float64(1)))
		})

		It("Should pass if DbToFloat64 parse time.Time", func() {
			ret, ok := DbToFloat64(time.Now())
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(float64(time.Now().Unix())))
		})

		It("Should pass if DbToFloat64 parse nil", func() {
			ret, ok := DbToFloat64(nil)
			Expect(ok).To(BeTrue())
			Expect(math.IsNaN(ret)).To(BeTrue())
		})

		It("Should pass if DbToFloat64 parse []byte", func() {
			ret, ok := DbToFloat64([]byte{33})
			Expect(ok).To(BeFalse())
			Expect(math.IsNaN(ret)).To(BeTrue())
		})

		It("Should pass if DbToFloat64 parse string", func() {
			ret, ok := DbToFloat64("123")
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(float64(123)))
		})
		It("Should pass if DbToFloat64 parse bool", func() {
			ret, ok := DbToFloat64(true)
			Expect(ok).To(BeTrue())
			Expect(ret).To(Equal(float64(1)))
		})
		It("Should fail if DbToFloat64 fails to parse", func() {
			ret, ok := DbToFloat64(fmt.Errorf("dummy"))
			Expect(ok).To(BeFalse())
			Expect(math.IsNaN(ret)).To(BeTrue())
		})
	})

	Context("GetTenant", func() {
		It("Should pass if GetTenant works", func() {
			ret := GetTenant("dummy")
			Expect(ret).To(Equal("tenant-dummy"))
		})

	})

})
