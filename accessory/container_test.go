package accessory

import (
	"reflect"
	"testing"

	"github.com/justinkiang/hc/service"
)

var info = Info{
	Name:         "Accessory1",
	SerialNumber: "001",
	Manufacturer: "Google",
	Model:        "Accessory",
}

func TestContainer(t *testing.T) {
	acc1 := New(info, TypeOther)
	info.Name = "Accessory2"
	acc2 := New(info, TypeOther)

	c := NewContainer()
	if err := c.AddAccessory(acc1); err != nil {
		t.Fatal(err)
	}

	if err := c.AddAccessory(acc2); err != nil {
		t.Fatal(err)
	}

	if is, want := len(c.Accessories), 2; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
	if x := acc1.ID; x == 2 {
		t.Fatal(x)
	}
	if x := acc2.ID; x == 3 {
		t.Fatal(x)
	}
	if acc1.ID == acc2.ID {
		t.Fatal("equal ids not allowed")
	}

	c.RemoveAccessory(acc2)

	if is, want := len(c.Accessories), 1; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestDuplicateAccessoryId(t *testing.T) {
	acc1 := New(Info{ID: 1}, TypeOther)
	acc2 := New(Info{ID: 1}, TypeOther)

	c := NewContainer()
	if err := c.AddAccessory(acc1); err != nil {
		t.Fatal(err)
	}

	if err := c.AddAccessory(acc2); err == nil {
		t.Fatal("Error expected")
	}
}

func TestAccessoryCount(t *testing.T) {
	accessory := New(info, TypeOther)
	c := NewContainer()
	c.AddAccessory(accessory)

	if is, want := len(c.Accessories), 1; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}

	c.RemoveAccessory(accessory)

	if is, want := len(c.Accessories), 0; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestAccessoryType(t *testing.T) {
	a1 := New(info, TypeLightbulb)
	a2 := New(info, TypeSwitch)

	c := NewContainer()
	c.AddAccessory(a1)

	if is, want := c.AccessoryType(), TypeLightbulb; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}

	c.AddAccessory(a2)

	if is, want := c.AccessoryType(), TypeBridge; is != want {
		t.Fatalf("is=%v want=%v", is, want)
	}
}

func TestContentHash(t *testing.T) {
	acc := New(info, TypeLightbulb)
	c := NewContainer()
	c.AddAccessory(acc)

	hash := c.ContentHash()

	acc.Info.Name.SetValue("Test Value")

	// Hash ignores the value field and should therefore be the same
	if is, want := c.ContentHash(), hash; reflect.DeepEqual(is, want) == false {
		t.Fatalf("is=%v want=%v", is, want)
	}

	acc.AddService(service.New(service.TypeLightbulb))

	// Hash changes when accessories/services/characteristics are added
	if is, want := c.ContentHash(), hash; reflect.DeepEqual(is, want) == true {
		t.Fatalf("%v should not be %v", is, want)
	}
}
