// Copyright (C) 2019-2020 Zilliz. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under the License

//
// Created by mike on 12/3/20.
//
#include "common/Types.h"
#include <knowhere/index/vector_index/helpers/IndexParameter.h>
#include "utils/EasyAssert.h"
#include <boost/bimap.hpp>
#include <boost/algorithm/string/case_conv.hpp>

namespace milvus {

using boost::algorithm::to_lower_copy;
namespace Metric = knowhere::Metric;
static auto map = [] {
    boost::bimap<std::string, MetricType> mapping;
    using pos = boost::bimap<std::string, MetricType>::value_type;
    mapping.insert(pos(to_lower_copy(std::string(Metric::L2)), MetricType::METRIC_L2));
    mapping.insert(pos(to_lower_copy(std::string(Metric::IP)), MetricType::METRIC_INNER_PRODUCT));
    mapping.insert(pos(to_lower_copy(std::string(Metric::JACCARD)), MetricType::METRIC_Jaccard));
    mapping.insert(pos(to_lower_copy(std::string(Metric::TANIMOTO)), MetricType::METRIC_Tanimoto));
    mapping.insert(pos(to_lower_copy(std::string(Metric::HAMMING)), MetricType::METRIC_Hamming));
    mapping.insert(pos(to_lower_copy(std::string(Metric::SUBSTRUCTURE)), MetricType::METRIC_Substructure));
    mapping.insert(pos(to_lower_copy(std::string(Metric::SUPERSTRUCTURE)), MetricType::METRIC_Superstructure));
    return mapping;
}();

MetricType
GetMetricType(const std::string& type_name) {
    auto real_name = to_lower_copy(type_name);
    AssertInfo(map.left.count(real_name), "metric type not found: (" + type_name + ")");
    return map.left.at(real_name);
}

}  // namespace milvus
