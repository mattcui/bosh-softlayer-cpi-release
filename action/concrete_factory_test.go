package action_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/maximilien/bosh-softlayer-cpi/action"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakecmd "github.com/cloudfoundry/bosh-agent/platform/commands/fakes"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	fakeslclient "github.com/maximilien/softlayer-go/client/fakes"

	bslcdisk "github.com/maximilien/bosh-softlayer-cpi/softlayer/disk"
	bslcstem "github.com/maximilien/bosh-softlayer-cpi/softlayer/stemcell"
	bslcvm "github.com/maximilien/bosh-softlayer-cpi/softlayer/vm"
)

var _ = Describe("concreteFactory", func() {
	var (
		softLayerClient *fakeslclient.FakeSoftLayerClient
		fs              *fakesys.FakeFileSystem
		cmdRunner       *fakesys.FakeCmdRunner
		compressor      *fakecmd.FakeCompressor
		logger          boshlog.Logger

		options = ConcreteFactoryOptions{
			StemcellsDir: "/tmp/stemcells",
		}

		factory Factory
	)

	var (
		agentEnvServiceFactory bslcvm.AgentEnvServiceFactory

		stemcellFinder bslcstem.Finder
		vmFinder       bslcvm.Finder
	)

	BeforeEach(func() {
		softLayerClient = fakeslclient.NewFakeSoftLayerClient("fake-username", "fake-api-key")
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		compressor = fakecmd.NewFakeCompressor()
		logger = boshlog.NewLogger(boshlog.LevelNone)

		factory = NewConcreteFactory(
			softLayerClient,
			options,
			logger,
		)
	})

	BeforeEach(func() {
		agentEnvServiceFactory = bslcvm.NewSoftLayerAgentEnvServiceFactory(softLayerClient, logger)

		stemcellFinder = bslcstem.NewSoftLayerFinder(softLayerClient, logger)

		vmFinder = bslcvm.NewSoftLayerFinder(
			softLayerClient,
			agentEnvServiceFactory,
			logger,
		)
	})

	Context("Stemcell methods", func() {
		It("create_stemcell", func() {
			action, err := factory.Create("create_stemcell")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewCreateStemcell(stemcellFinder)))
		})

		It("delete_stemcell", func() {
			action, err := factory.Create("delete_stemcell")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewDeleteStemcell(stemcellFinder)))
		})
	})

	Context("VM methods", func() {
		It("create_vm", func() {
			vmCreator := bslcvm.NewSoftLayerCreator(
				softLayerClient,
				agentEnvServiceFactory,
				options.Agent,
				logger,
			)

			action, err := factory.Create("create_vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewCreateVM(stemcellFinder, vmCreator)))
		})

		It("delete_vm", func() {
			action, err := factory.Create("delete_vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewDeleteVM(vmFinder)))
		})

		It("has_vm", func() {
			action, err := factory.Create("has_vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewHasVM(vmFinder)))
		})

		It("reboot_vm", func() {
			action, err := factory.Create("reboot_vm")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewRebootVM(vmFinder)))
		})

		It("set_vm_metadata", func() {
			action, err := factory.Create("set_vm_metadata")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewSetVMMetadata(vmFinder)))
		})

		It("configure_networks", func() {
			action, err := factory.Create("configure_networks")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewConfigureNetworks(vmFinder)))
		})
	})

	//TODO: implement when disk methods and actions are added
	XContext("Disk methods", func() {
		It("create_disk", func() {
			diskCreator := bslcdisk.NewFSCreator(
				"/tmp/disks",
				fs,
				cmdRunner,
				logger,
			)

			action, err := factory.Create("create_disk")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(Equal(NewCreateDisk(diskCreator)))
		})

		It("delete_disk", func() {
			action, err := factory.Create("delete_disk")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("attach_disk", func() {
			action, err := factory.Create("attach_disk")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("detach_disk", func() {
			action, err := factory.Create("detach_disk")
			Expect(err).ToNot(HaveOccurred())
			Expect(action).To(BeNil())
		})
	})

	Context("Unsupported methods", func() {
		It("returns error because CPI machine is not self-aware if action is current_vm_id", func() {
			action, err := factory.Create("current_vm_id")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because snapshotting is not implemented if action is snapshot_disk", func() {
			action, err := factory.Create("snapshot_disk")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because snapshotting is not implemented if action is delete_snapshot", func() {
			action, err := factory.Create("delete_snapshot")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error since CPI should not keep state if action is get_disks", func() {
			action, err := factory.Create("get_disks")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})

		It("returns error because ping is not official CPI method if action is ping", func() {
			action, err := factory.Create("ping")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})
	})

	Context("Misc", func() {
		It("returns error if action cannot be created", func() {
			action, err := factory.Create("fake-unknown-action")
			Expect(err).To(HaveOccurred())
			Expect(action).To(BeNil())
		})
	})
})
