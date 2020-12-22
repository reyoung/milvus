package master

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/zilliztech/milvus-distributed/internal/proto/commonpb"
	"github.com/zilliztech/milvus-distributed/internal/proto/internalpb"
	"github.com/zilliztech/milvus-distributed/internal/proto/masterpb"
	"github.com/zilliztech/milvus-distributed/internal/proto/schemapb"
	"github.com/zilliztech/milvus-distributed/internal/proto/servicepb"
	"github.com/zilliztech/milvus-distributed/internal/util/typeutil"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
)

func TestMaster_Partition(t *testing.T) {
	Init()
	refreshMasterAddress()

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	etcdAddr := Params.EtcdAddress
	etcdCli, err := clientv3.New(clientv3.Config{Endpoints: []string{etcdAddr}})
	assert.Nil(t, err)
	_, err = etcdCli.Delete(ctx, "/test/root", clientv3.WithPrefix())
	assert.Nil(t, err)

	Params = ParamTable{
		Address: Params.Address,
		Port:    Params.Port,

		EtcdAddress:   Params.EtcdAddress,
		MetaRootPath:  "/test/root/meta",
		KvRootPath:    "/test/root/kv",
		PulsarAddress: Params.PulsarAddress,

		ProxyIDList:     []typeutil.UniqueID{1, 2},
		WriteNodeIDList: []typeutil.UniqueID{3, 4},

		TopicNum:                    5,
		QueryNodeNum:                3,
		SoftTimeTickBarrierInterval: 300,

		// segment
		SegmentSize:           536870912 / 1024 / 1024,
		SegmentSizeFactor:     0.75,
		DefaultRecordSize:     1024,
		MinSegIDAssignCnt:     1048576 / 1024,
		MaxSegIDAssignCnt:     Params.MaxSegIDAssignCnt,
		SegIDAssignExpiration: 2000,

		// msgChannel
		ProxyTimeTickChannelNames:     []string{"proxy1", "proxy2"},
		WriteNodeTimeTickChannelNames: []string{"write3", "write4"},
		InsertChannelNames:            []string{"dm0", "dm1"},
		K2SChannelNames:               []string{"k2s0", "k2s1"},
		QueryNodeStatsChannelName:     "statistic",
		MsgChannelSubName:             Params.MsgChannelSubName,

		MaxPartitionNum:     int64(4096),
		DefaultPartitionTag: "_default",
	}

	svr, err := CreateServer(ctx)
	assert.Nil(t, err)
	err = svr.Run(int64(Params.Port))
	assert.Nil(t, err)

	conn, err := grpc.DialContext(ctx, Params.Address, grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer conn.Close()

	cli := masterpb.NewMasterClient(conn)
	sch := schemapb.CollectionSchema{
		Name:        "col1",
		Description: "test collection",
		AutoID:      false,
		Fields: []*schemapb.FieldSchema{
			{
				Name:        "col1_f1",
				Description: "test collection filed 1",
				DataType:    schemapb.DataType_VECTOR_FLOAT,
				TypeParams: []*commonpb.KeyValuePair{
					{
						Key:   "col1_f1_tk1",
						Value: "col1_f1_tv1",
					},
					{
						Key:   "col1_f1_tk2",
						Value: "col1_f1_tv2",
					},
				},
				IndexParams: []*commonpb.KeyValuePair{
					{
						Key:   "col1_f1_ik1",
						Value: "col1_f1_iv1",
					},
					{
						Key:   "col1_f1_ik2",
						Value: "col1_f1_iv2",
					},
				},
			},
			{
				Name:        "col1_f2",
				Description: "test collection filed 2",
				DataType:    schemapb.DataType_VECTOR_BINARY,
				TypeParams: []*commonpb.KeyValuePair{
					{
						Key:   "col1_f2_tk1",
						Value: "col1_f2_tv1",
					},
					{
						Key:   "col1_f2_tk2",
						Value: "col1_f2_tv2",
					},
				},
				IndexParams: []*commonpb.KeyValuePair{
					{
						Key:   "col1_f2_ik1",
						Value: "col1_f2_iv1",
					},
					{
						Key:   "col1_f2_ik2",
						Value: "col1_f2_iv2",
					},
				},
			},
		},
	}
	schemaBytes, err := proto.Marshal(&sch)
	assert.Nil(t, err)

	createCollectionReq := internalpb.CreateCollectionRequest{
		MsgType:   internalpb.MsgType_kCreatePartition,
		ReqID:     1,
		Timestamp: 1,
		ProxyID:   1,
		Schema:    &commonpb.Blob{Value: schemaBytes},
	}
	st, _ := cli.CreateCollection(ctx, &createCollectionReq)
	assert.NotNil(t, st)
	assert.Equal(t, commonpb.ErrorCode_SUCCESS, st.ErrorCode)

	createPartitionReq := internalpb.CreatePartitionRequest{
		MsgType:       internalpb.MsgType_kCreatePartition,
		ReqID:         1,
		Timestamp:     2,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}
	st, _ = cli.CreatePartition(ctx, &createPartitionReq)
	assert.NotNil(t, st)
	assert.Equal(t, commonpb.ErrorCode_SUCCESS, st.ErrorCode)

	createPartitionReq = internalpb.CreatePartitionRequest{
		MsgType:       internalpb.MsgType_kCreatePartition,
		ReqID:         1,
		Timestamp:     1,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}
	st, _ = cli.CreatePartition(ctx, &createPartitionReq)
	assert.NotNil(t, st)
	assert.Equal(t, commonpb.ErrorCode_UNEXPECTED_ERROR, st.ErrorCode)

	createPartitionReq = internalpb.CreatePartitionRequest{
		MsgType:       internalpb.MsgType_kCreatePartition,
		ReqID:         1,
		Timestamp:     3,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition2"},
	}
	st, _ = cli.CreatePartition(ctx, &createPartitionReq)
	assert.NotNil(t, st)
	assert.Equal(t, commonpb.ErrorCode_SUCCESS, st.ErrorCode)

	collMeta, err := svr.metaTable.GetCollectionByName(sch.Name)
	assert.Nil(t, err)
	t.Logf("collection id = %d", collMeta.ID)
	assert.Equal(t, collMeta.CreateTime, uint64(1))
	assert.Equal(t, collMeta.Schema.Name, "col1")
	assert.Equal(t, collMeta.Schema.AutoID, false)
	assert.Equal(t, len(collMeta.Schema.Fields), 2)
	assert.Equal(t, collMeta.Schema.Fields[0].Name, "col1_f1")
	assert.Equal(t, collMeta.Schema.Fields[1].Name, "col1_f2")
	assert.Equal(t, collMeta.Schema.Fields[0].DataType, schemapb.DataType_VECTOR_FLOAT)
	assert.Equal(t, collMeta.Schema.Fields[1].DataType, schemapb.DataType_VECTOR_BINARY)
	assert.Equal(t, len(collMeta.Schema.Fields[0].TypeParams), 2)
	assert.Equal(t, len(collMeta.Schema.Fields[0].IndexParams), 2)
	assert.Equal(t, len(collMeta.Schema.Fields[1].TypeParams), 2)
	assert.Equal(t, len(collMeta.Schema.Fields[1].IndexParams), 2)
	assert.Equal(t, collMeta.Schema.Fields[0].TypeParams[0].Key, "col1_f1_tk1")
	assert.Equal(t, collMeta.Schema.Fields[0].TypeParams[1].Key, "col1_f1_tk2")
	assert.Equal(t, collMeta.Schema.Fields[0].TypeParams[0].Value, "col1_f1_tv1")
	assert.Equal(t, collMeta.Schema.Fields[0].TypeParams[1].Value, "col1_f1_tv2")
	assert.Equal(t, collMeta.Schema.Fields[0].IndexParams[0].Key, "col1_f1_ik1")
	assert.Equal(t, collMeta.Schema.Fields[0].IndexParams[1].Key, "col1_f1_ik2")
	assert.Equal(t, collMeta.Schema.Fields[0].IndexParams[0].Value, "col1_f1_iv1")
	assert.Equal(t, collMeta.Schema.Fields[0].IndexParams[1].Value, "col1_f1_iv2")

	assert.Equal(t, collMeta.Schema.Fields[1].TypeParams[0].Key, "col1_f2_tk1")
	assert.Equal(t, collMeta.Schema.Fields[1].TypeParams[1].Key, "col1_f2_tk2")
	assert.Equal(t, collMeta.Schema.Fields[1].TypeParams[0].Value, "col1_f2_tv1")
	assert.Equal(t, collMeta.Schema.Fields[1].TypeParams[1].Value, "col1_f2_tv2")
	assert.Equal(t, collMeta.Schema.Fields[1].IndexParams[0].Key, "col1_f2_ik1")
	assert.Equal(t, collMeta.Schema.Fields[1].IndexParams[1].Key, "col1_f2_ik2")
	assert.Equal(t, collMeta.Schema.Fields[1].IndexParams[0].Value, "col1_f2_iv1")
	assert.Equal(t, collMeta.Schema.Fields[1].IndexParams[1].Value, "col1_f2_iv2")

	//assert.Equal(t, collMeta.PartitionTags[0], "partition1")
	//assert.Equal(t, collMeta.PartitionTags[1], "partition2")
	assert.ElementsMatch(t, []string{"_default", "partition1", "partition2"}, collMeta.PartitionTags)

	showPartitionReq := internalpb.ShowPartitionRequest{
		MsgType:        internalpb.MsgType_kShowPartitions,
		ReqID:          1,
		Timestamp:      4,
		ProxyID:        1,
		CollectionName: &servicepb.CollectionName{CollectionName: "col1"},
	}

	stringList, err := cli.ShowPartitions(ctx, &showPartitionReq)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"_default", "partition1", "partition2"}, stringList.Values)

	showPartitionReq = internalpb.ShowPartitionRequest{
		MsgType:        internalpb.MsgType_kShowPartitions,
		ReqID:          1,
		Timestamp:      3,
		ProxyID:        1,
		CollectionName: &servicepb.CollectionName{CollectionName: "col1"},
	}

	stringList, _ = cli.ShowPartitions(ctx, &showPartitionReq)
	assert.NotNil(t, stringList)
	assert.Equal(t, commonpb.ErrorCode_UNEXPECTED_ERROR, stringList.Status.ErrorCode)

	hasPartitionReq := internalpb.HasPartitionRequest{
		MsgType:       internalpb.MsgType_kHasPartition,
		ReqID:         1,
		Timestamp:     5,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}

	hasPartition, err := cli.HasPartition(ctx, &hasPartitionReq)
	assert.Nil(t, err)
	assert.True(t, hasPartition.Value)

	hasPartitionReq = internalpb.HasPartitionRequest{
		MsgType:       internalpb.MsgType_kHasPartition,
		ReqID:         1,
		Timestamp:     4,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}

	hasPartition, _ = cli.HasPartition(ctx, &hasPartitionReq)
	assert.NotNil(t, hasPartition)
	assert.Equal(t, commonpb.ErrorCode_UNEXPECTED_ERROR, stringList.Status.ErrorCode)

	hasPartitionReq = internalpb.HasPartitionRequest{
		MsgType:       internalpb.MsgType_kHasPartition,
		ReqID:         1,
		Timestamp:     6,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition3"},
	}

	hasPartition, err = cli.HasPartition(ctx, &hasPartitionReq)
	assert.Nil(t, err)
	assert.False(t, hasPartition.Value)

	deletePartitionReq := internalpb.DropPartitionRequest{
		MsgType:       internalpb.MsgType_kDropPartition,
		ReqID:         1,
		Timestamp:     7,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition2"},
	}

	st, err = cli.DropPartition(ctx, &deletePartitionReq)
	assert.Nil(t, err)
	assert.Equal(t, commonpb.ErrorCode_SUCCESS, st.ErrorCode)

	deletePartitionReq = internalpb.DropPartitionRequest{
		MsgType:       internalpb.MsgType_kDropPartition,
		ReqID:         1,
		Timestamp:     6,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition2"},
	}

	st, _ = cli.DropPartition(ctx, &deletePartitionReq)
	assert.NotNil(t, st)
	assert.Equal(t, commonpb.ErrorCode_UNEXPECTED_ERROR, st.ErrorCode)

	hasPartitionReq = internalpb.HasPartitionRequest{
		MsgType:       internalpb.MsgType_kHasPartition,
		ReqID:         1,
		Timestamp:     8,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition2"},
	}

	hasPartition, err = cli.HasPartition(ctx, &hasPartitionReq)
	assert.Nil(t, err)
	assert.False(t, hasPartition.Value)

	describePartitionReq := internalpb.DescribePartitionRequest{
		MsgType:       internalpb.MsgType_kDescribePartition,
		ReqID:         1,
		Timestamp:     9,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}

	describePartition, err := cli.DescribePartition(ctx, &describePartitionReq)
	assert.Nil(t, err)
	assert.Equal(t, &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"}, describePartition.Name)

	describePartitionReq = internalpb.DescribePartitionRequest{
		MsgType:       internalpb.MsgType_kDescribePartition,
		ReqID:         1,
		Timestamp:     8,
		ProxyID:       1,
		PartitionName: &servicepb.PartitionName{CollectionName: "col1", Tag: "partition1"},
	}

	describePartition, _ = cli.DescribePartition(ctx, &describePartitionReq)
	assert.Equal(t, commonpb.ErrorCode_UNEXPECTED_ERROR, describePartition.Status.ErrorCode)

	svr.Close()
}