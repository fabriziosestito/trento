import React, { useEffect, useState, useRef } from 'react';
import ReactDOM from 'react-dom';
import { get } from 'axios';
import Tags from "@yaireo/tagify/dist/react.tagify"

const HostList = () => {
  const [hosts, setHosts] = useState([]);

  useEffect(() => {
    get(`https://6189698fd0821900178d79b2.mockapi.io/api/hosts`).then(({ data }) => {

      setHosts(data);
    });
  }, []);


  return (
    <div className="table-responsive">
      <table className="table eos-table">
        <thead>
          <tr>
            <th scope='col'></th>
            <th scope='col'>Name</th>
            <th scope='col'>Address</th>
            <th scope='col'>Cloud provider</th>
            <th scope='col'>Cluster</th>
            <th scope='col'>System</th>
            <th scope='col'>Agent version</th>
            <th scope='col'>Tags</th>
          </tr>
        </thead>
        <tbody>
          {hosts.map((host) => (
            <tr key={host.id}>
              <td className="row-status">
                health
              </td>
              <td>
                {host.name}
              </td>
              <td>
                {host.ip_addresses}
              </td>
              <td>
                {host.cloud_provider}
              </td>
              <td>
                {host.cluster}
              </td>
              <td>
                {host.system}
              </td>
              <td>
                {host.agent_version}
              </td>
              <TagRow
                tags={host.tags}
              />
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

function TagRow({ tags }) {
  const tagifyRef = useRef()
  const settings = {
    whitelist: [],
    editTags: false,
    pattern: /^[0-9A-Za-z\s\-_]+$/,
    dropdown: {
      maxItems: 20,
      enabled: 1,
      closeOnSelect: false,
      placeAbove: false,
      classname: 'tags-look',
    }
  };

  return (
    <td onClick={() => {
      tagifyRef.current.addEmptyTag()
    }}>
      <Tags
        settings={settings}
        tagifyRef={tagifyRef}
        defaultValue={tags}
      />
    </td>
  );
}

ReactDOM.render(
  <HostList />,
  document.getElementById('host-list')
);
