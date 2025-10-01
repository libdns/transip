package transip

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/libdns/libdns"
	"github.com/pbergman/provider"
)

type DNSProvider interface {
	libdns.RecordAppender
	libdns.RecordDeleter
	libdns.RecordGetter
	libdns.RecordSetter
}

func newProvider() DNSProvider {

	p := &Provider{
		AuthLogin:  os.Getenv("LOGIN"),
		PrivateKey: os.Getenv("KEY"),
	}

	if _, ok := os.LookupEnv("DEBUG"); ok {
		p.Debug = true
	}

	return p
}

func printRecords(t *testing.T, records []libdns.Record, invalid libdns.Record) {

	var buf = new(bytes.Buffer)
	var writer = tabwriter.NewWriter(buf, 0, 4, 2, ' ', tabwriter.Debug)
	var isWritten = false

	for _, record := range records {
		var rr = record.RR()

		if invalid == nil {
			_, _ = fmt.Fprintf(writer, " %s\t %s\t %s\t %s\n", rr.Name, rr.TTL, rr.Type, rr.Data)
		} else {
			var prefix = "  "
			if record.RR().Type == invalid.RR().Type && record.RR().Data == invalid.RR().Data && record.RR().Name == invalid.RR().Name {
				prefix = "- "
				isWritten = true
			}
			_, _ = fmt.Fprintf(writer, "%s%s\t %s\t %s\t %s\n", prefix, rr.Name, rr.TTL, rr.Type, rr.Data)
		}
	}

	if false == isWritten && nil != invalid {
		_, _ = fmt.Fprintf(writer, "%s%s\t %s\t %s\t %s\n", "+ ", invalid.RR().Name, invalid.RR().TTL, invalid.RR().Type, invalid.RR().Data)
	}

	_ = writer.Flush()

	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		t.Log(scanner.Text())
	}
}

func getZones(p DNSProvider, t *testing.T) []string {

	if v, ok := os.LookupEnv("ZONE"); ok {
		return strings.Split(v, ",")
	}

	if o, ok := p.(libdns.ZoneLister); ok {
		zones, err := o.ListZones(context.Background())

		if err != nil {
			t.Fatalf("ListZones failed: %v", err)
		}

		var ret = make([]string, len(zones))

		for idx, zone := range zones {
			ret[idx] = zone.Name
		}

		return ret
	}

	return []string{"example.com"}
}

func TestProvider_ListZones(t *testing.T) {
	var p = newProvider()

	if x, ok := p.(libdns.ZoneLister); ok {
		zones, err := x.ListZones(context.Background())

		if err != nil {
			t.Fatalf("ListZones failed: %v", err)
		}

		t.Log("Available zones:")

		for _, zone := range zones {
			t.Logf("%#+v", zone)
		}
	} else {
		t.Skipf("ListZones not implemented.")
	}
}

func TestProvider_GetRecords(t *testing.T) {
	var p = newProvider()

	for _, zone := range getZones(p, t) {
		records, err := p.GetRecords(context.Background(), zone)

		if err != nil {
			t.Fatalf("GetRecords failed: %v", err)
		}

		printRecords(t, records, nil)
	}

}

func TestProvider_AppendRecords(t *testing.T) {

	var records = []libdns.Record{
		libdns.TXT{
			Name: "_libdns_test_append_records_",
			Text: "a",
		},
		libdns.TXT{
			Name: "_libdns_test_append_records_",
			Text: "b",
		},
	}

	var p = newProvider()

	for _, zone := range getZones(p, t) {

		out, err := p.AppendRecords(context.Background(), zone, records)

		if err != nil {
			t.Fatalf("AppendRecords failed: %v", err)
		}

		t.Logf("Appended successfully %d rocords to zone %s", len(records), zone)

		for _, record := range provider.RecordIterator(&records) {
			if false == provider.IsInList(&record, &out) {
				t.Fatalf("AppendRecords returned unexpected record, expecting %+v, to be in list %+v", record, out)
			}
		}

		printRecords(t, records, nil)

		if _, err := p.AppendRecords(context.Background(), zone, records); err == nil {
			t.Fatalf("AppendRecords should have failed but didn't")
		}

		_, _ = p.DeleteRecords(context.Background(), zone, records)
	}
}

func TestProvider_SetRecords(t *testing.T) {

	var p = newProvider()
	var name = "libdns.test.set.records"

	for _, zone := range getZones(p, t) {
		t.Run("SetRecords Example 1", func(t *testing.T) {
			setRecordsExample1(t, p, zone, name)
		})
		t.Run("SetRecords Example 2", func(t *testing.T) {
			setRecordsExample2(t, p, zone, name)
		})
	}
}

func setRecordsExample1(t *testing.T, p DNSProvider, zone, name string) {

	var original = []libdns.Record{
		libdns.Address{Name: name, IP: netip.MustParseAddr("192.0.2.1")},
		libdns.Address{Name: name, IP: netip.MustParseAddr("192.0.2.2")},
		libdns.TXT{Name: name, Text: "hello world"},
	}

	var input = []libdns.Record{
		libdns.Address{Name: name, IP: netip.MustParseAddr("192.0.2.3")},
	}

	out, err := p.AppendRecords(context.Background(), zone, original)

	if err != nil {
		t.Fatalf("AppendRecords failed: %v", err)
	}

	t.Logf("Set test records for zone \"%s\":", zone)

	printRecords(t, out, nil)

	// make sure we delete all records even on failure
	defer p.DeleteRecords(context.Background(), zone, []libdns.Record{
		libdns.Address{Name: name},
		libdns.TXT{Name: name},
	})

	ret, err := p.SetRecords(context.Background(), zone, input)

	if err != nil {
		t.Fatalf("SetRecords failed: %v", err)
	}

	if len(ret) != 1 {
		t.Fatalf("SetRecords should have returned 1 record")
	}

	curr, err := p.GetRecords(context.Background(), zone)

	if err != nil {
		t.Fatalf("GetRecords failed: %v", err)
	}

	t.Log("Current records zone:")
	printRecords(t, curr, nil)

	var shouldNotExist = original[:2]

	for invalid, record := range provider.RecordIterator(&shouldNotExist) {
		if provider.IsInList(&record, &curr) {
			printRecords(t, curr, *invalid)
			t.Fatal("AppendRecords returned unexpected records")
		}
	}

	var shouldExist = append(original[2:], input[0])

	for invalid, record := range provider.RecordIterator(&shouldExist) {
		if false == provider.IsInList(&record, &curr) {
			printRecords(t, curr, *invalid)
			t.Fatal("AppendRecords returned unexpected records")
		}
	}
}

func setRecordsExample2(t *testing.T, p DNSProvider, zone, name string) {

	var original = []libdns.Record{
		libdns.Address{Name: "alpha." + name, IP: netip.MustParseAddr("2001:db8::1")},
		libdns.Address{Name: "alpha." + name, IP: netip.MustParseAddr("2001:db8::2")},
		libdns.Address{Name: "beta." + name, IP: netip.MustParseAddr("2001:db8::3")},
		libdns.Address{Name: "beta." + name, IP: netip.MustParseAddr("2001:db8::4")},
	}

	var input = []libdns.Record{
		libdns.Address{Name: "alpha." + name, IP: netip.MustParseAddr("2001:db8::1")},
		libdns.Address{Name: "alpha." + name, IP: netip.MustParseAddr("2001:db8::2")},
		libdns.Address{Name: "alpha." + name, IP: netip.MustParseAddr("2001:db8::5")},
	}

	out, err := p.AppendRecords(context.Background(), zone, original)

	if err != nil {
		t.Fatalf("AppendRecords failed: %v", err)
	}

	t.Logf("Set test records for zone \"%s\":", zone)

	printRecords(t, out, nil)

	// make sure we delete all records even on failure
	defer p.DeleteRecords(context.Background(), zone, []libdns.Record{
		libdns.RR{Name: "alpha." + name, Type: "AAAA"},
		libdns.RR{Name: "beta." + name, Type: "AAAA"},
	})

	ret, err := p.SetRecords(context.Background(), zone, input)

	if err != nil {
		t.Fatalf("SetRecords failed: %v", err)
	}

	if len(ret) != 1 {
		t.Fatalf("SetRecords should have returned 1 record")
	}

	curr, err := p.GetRecords(context.Background(), zone)

	if err != nil {
		t.Fatalf("GetRecords failed: %v", err)
	}

	t.Log("Current records zone:")
	printRecords(t, curr, nil)

	var shouldExist = append(original, input[2])

	for invalid, record := range provider.RecordIterator(&shouldExist) {
		if false == provider.IsInList(&record, &curr) {
			printRecords(t, curr, *invalid)
			t.Fatal("AppendRecords returned unexpected records")
		}
	}
}

func TestProvider_DeleteRecords(t *testing.T) {

	var p = newProvider()
	var name = "libdns.test.rm.records"

	for _, zone := range getZones(p, t) {
		var records = []libdns.Record{
			libdns.Address{Name: name, IP: netip.MustParseAddr("2001:db8::1")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("2001:db8::2")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("2001:db8::3")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("2001:db8::4")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("2001:db8::5")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.1")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.2")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.3")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.4")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.5")},
			libdns.Address{Name: name, IP: netip.MustParseAddr("127.0.0.6")},
		}

		out, err := p.SetRecords(context.Background(), zone, records)

		if err != nil {
			t.Fatalf("SetRecords failed: %v", err)
		}

		t.Logf("Set following test records for zone \"%s\":", zone)

		printRecords(t, out, nil)

		var toRemove = records[:2]

		removed, err := p.DeleteRecords(context.Background(), zone, toRemove)

		if err != nil {
			t.Fatalf("DeleteRecords failed: %v", err)
		}

		for _, x := range provider.RecordIterator(&removed) {
			if false == provider.IsInList(&x, &toRemove) {
				printRecords(t, toRemove, x)
				t.Fatal("DeleteRecords returned unexpected records")
			}
		}

		t.Log("Deleted records:")
		printRecords(t, removed, nil)

		curr, err := p.GetRecords(context.Background(), zone)

		if err != nil {
			t.Fatalf("GetRecords failed: %v", err)
		}

		for _, x := range provider.RecordIterator(&toRemove) {
			if provider.IsInList(&x, &curr) {
				printRecords(t, curr, x)
				t.Fatal("DeleteRecords returned unexpected records")
			}
		}

		removed, err = p.DeleteRecords(context.Background(), zone, []libdns.Record{
			libdns.RR{Name: name, Type: "AAAA"},
			libdns.RR{Name: name, Type: "A"},
		})

		if err != nil {
			t.Fatalf("DeleteRecords failed: %v", err)
		}
		t.Log("Deleted records:")
		printRecords(t, removed, nil)

		if len(removed) != 9 {
			t.Fatalf("DeleteRecords returned invalid count of records: expecting 9 got %d", len(removed))
		}
	}
}
